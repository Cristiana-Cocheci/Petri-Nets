package src

import (
	"fmt"
	r "math/rand"
	"slices"
)

type WorkCluster struct {
	// transitions in a work cluster are fired sequentially
	Id          int
	Places      map[string]struct{} // set of place nodes
	Transitions map[string]struct{} // set of transition nodes
	FireChannel chan struct{}       // message channel to fire if updates occur in work cluster
}

func (workCluster *WorkCluster) PrintWorkCluster() {
	fmt.Printf("Work Cluster nr. %d: \n", workCluster.Id)
	fmt.Print("\tPlaces: ")
	for place := range workCluster.Places {
		fmt.Printf("%s ", place)
	}
	fmt.Print("\n\tTransitions: ")
	for transition := range workCluster.Transitions {
		fmt.Printf("%s ", transition)
	}
	fmt.Print("\n")
}

func (workCluster *WorkCluster) NewWorkCluster(id *int) {
	workCluster.Id = *id
	workCluster.Places = make(map[string]struct{})
	workCluster.Transitions = make(map[string]struct{})
	workCluster.FireChannel = make(chan struct{}, 1)
	*id++
}

func (workCluster *WorkCluster) AddPlace(place string, net *Net) {
	_, contained := workCluster.Places[place]
	if contained {
		return
	}
	workCluster.Places[place] = struct{}{}
	// Also add all new transitions that have that place as its input
	for _, transitionEdge := range net.InEdges[place] {
		workCluster.AddTransition(transitionEdge.Transition, net)
	}
}

func (workCluster *WorkCluster) AddTransition(transition string, net *Net) {
	_, contained := workCluster.Transitions[transition]
	if contained {
		return
	}
	workCluster.Transitions[transition] = struct{}{}
	// Also add all new places that this transition has as its input
	for _, placeEdge := range net.ReverseInEdges[transition] {
		workCluster.AddPlace(placeEdge.Place, net)
	}
}

func (net *Net) SplitNet() {
	var workClusters []WorkCluster
	var currentCluster WorkCluster
	visitedPlaces := make(map[string]struct{})
	currentId := 0
	for place := range net.Places {
		_, visited := visitedPlaces[place]
		if visited {
			continue
		}
		// Create a new work cluster for the free place
		currentCluster.NewWorkCluster(&currentId)
		// Add place to the new work cluster (along with all connected transitions and neighbouring places)
		currentCluster.AddPlace(place, net)
		// Mark all places from cluster as visited
		for p := range currentCluster.Places {
			visitedPlaces[p] = struct{}{}
		}
		// Add current work cluster to list
		workClusters = append(workClusters, currentCluster)
	}
	net.WorkClusters = workClusters
	net.MatchWorkClustersToTriggeringTransitions()
}

func (net *Net) MatchWorkClustersToTriggeringTransitions() {
	for transition := range net.Transitions {
		for _, edge := range net.OutEdges[transition] {
			for _, workCluster := range net.WorkClusters {
				_, placeIsInCluster := workCluster.Places[edge.Place]
				if placeIsInCluster && !slices.Contains(net.Transitions[transition], workCluster.Id) {
					net.Transitions[transition] = append(net.Transitions[transition], workCluster.Id)
				}
			}
		}
	}
	for transition := range net.Transitions {
		fmt.Printf("Transition %s triggers work clusters: %v\n", transition, net.Transitions[transition])
	}

}

func (workCluster *WorkCluster) checkFire(net *Net) {
	defer func() {
		if recover() != nil {
			// tried to send on closed channel
			fmt.Println("Goroutine timeout")
		}
	}()
	for range workCluster.FireChannel {
		// Wait for a transition corresponding to the Work Cluster to update its tokens
		var enabledTransitions []string

		for transition := range workCluster.Transitions {
			// check if transition can be fired (enough tokens collected in incoming places)
			canFire := true

			for _, pe := range net.ReverseInEdges[transition] {
				weight := pe.Weight
				incomingPlaces := net.Places[pe.Place]
				if incomingPlaces.tokens < weight {
					canFire = false
					break
				}
			}

			if canFire {
				enabledTransitions = append(enabledTransitions, transition)
			}
		}
		// For non-deterministic behaviour, pick a random transition to fire
		if len(enabledTransitions) > 0 {
			select {
			case workCluster.FireChannel <- struct{}{}: // send signal to work cluster if not empty
			default: // do nothing if channel is full
			}
			net.Fire(enabledTransitions[r.Int()%len(enabledTransitions)])
		}
	}
}
