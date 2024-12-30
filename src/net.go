package src

import (
	"fmt"
	"sync"
)

type WeightedTransitionEdge struct {
	// weight specifies number of tokens required to fire place->transition
	Transition string
	Weight     int
}

func (edge *WeightedTransitionEdge) PrintEdge() {
	fmt.Printf("\t\tTransition %s: Weight %d\n", edge.Transition, edge.Weight)
}

type WeightedPlaceEdge struct {
	// weight specifies number of tokens required to fire transition -> place
	// weights must not be symmetric to the WeightedTransitionEdge
	// tokens can be "lost" or "multiplied" in the transition node
	Place  string
	Weight int
}

func (edge *WeightedPlaceEdge) PrintEdge() {
	fmt.Printf("\t\tPlace %s: Weight %d\n", edge.Place, edge.Weight)
}

type Place struct {
	// place is a node in the Petri net
	tokens int         // number of tokens in the place at a time t
	mutex  *sync.Mutex // mutex to lock the place while adding/removing tokens
}

type Net struct {
	Places         map[string]*Place                   // holds number of current tokens
	Transitions    map[string]struct{}                 // set of transition nodes
	InEdges        map[string][]WeightedTransitionEdge // places -> transitions
	ReverseInEdges map[string][]WeightedPlaceEdge      // reverse directed graph of InEdges
	OutEdges       map[string][]WeightedPlaceEdge      // transitions -> places
	WorkClusters   []WorkCluster
}

func (net *Net) NewNet() {
	net.Places = make(map[string]*Place)
	net.Transitions = make(map[string]struct{})
	net.InEdges = make(map[string][]WeightedTransitionEdge)
	net.ReverseInEdges = make(map[string][]WeightedPlaceEdge)
	net.OutEdges = make(map[string][]WeightedPlaceEdge)
}

func (net *Net) AddPlace(place string, tokens int) {
	net.Places[place] = &Place{tokens, &sync.Mutex{}}
}

func (net *Net) AddTransition(transition string) {
	net.Transitions[transition] = struct{}{}
}

func (net *Net) AddEdge(from string, to string, weight int) {
	_, fromIsPlace := net.Places[from]
	_, fromIsTransition := net.Transitions[from]
	_, toIsPlace := net.Places[to]
	_, toIsTransition := net.Transitions[to]

	if fromIsPlace && toIsTransition {
		net.InEdges[from] = append(net.InEdges[from], WeightedTransitionEdge{Transition: to, Weight: weight})
		net.ReverseInEdges[to] = append(net.ReverseInEdges[to], WeightedPlaceEdge{Place: from, Weight: weight})
	} else if fromIsTransition && toIsPlace {
		net.OutEdges[from] = append(net.OutEdges[from], WeightedPlaceEdge{Place: to, Weight: weight})
	} else {
		PrintError(fmt.Errorf("invalid edge from %s to %s", from, to))
	}
}

func (net *Net) Run() {
	net.SplitNet()
	for _, workCluster := range net.WorkClusters {
		workCluster.PrintWorkCluster()
		go workCluster.checkFire(net)
	}
}

func (net *Net) Fire(transition string) {
	fmt.Printf("Fired transition %s \n", transition)
	// remove tokens from incoming places
	for _, pe := range net.ReverseInEdges[transition] {
		net.Places[pe.Place].mutex.Lock()
		net.Places[pe.Place].tokens -= pe.Weight
		net.Places[pe.Place].mutex.Unlock()
	}
	// add tokens to outgoing places
	for _, pe := range net.OutEdges[transition] {
		net.Places[pe.Place].mutex.Lock()
		net.Places[pe.Place].tokens += pe.Weight
		net.Places[pe.Place].mutex.Unlock()
	}
}

func (net *Net) PrintNet() {
	fmt.Println("Places:")
	for place, p := range net.Places {
		fmt.Printf("\t%s: %d\n", place, p.tokens)
	}
	fmt.Println("Transitions:")
	for transition := range net.Transitions {
		fmt.Printf("\t%s\n", transition)
	}
	fmt.Println("Edges:")
	for place, edges := range net.InEdges {
		fmt.Printf("\tPlace %s -> \n", place)
		for _, edge := range edges {
			edge.PrintEdge()
		}
	}
	for transition, edges := range net.OutEdges {
		fmt.Printf("\tTransition %s -> \n", transition)
		for _, edge := range edges {
			edge.PrintEdge()
		}
	}
}
