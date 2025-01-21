package src

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
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
	Places            map[string]*Place                   // holds number of current tokens
	Transitions       map[string][]int                    // set of transition nodes -> ids of work clusters they trigger
	InEdges           map[string][]WeightedTransitionEdge // places -> transitions
	ReverseInEdges    map[string][]WeightedPlaceEdge      // reverse directed graph of InEdges
	OutEdges          map[string][]WeightedPlaceEdge      // transitions -> places
	WorkClusters      []WorkCluster                       // no. of workclusters = no. of useful threads
	ClosingChannel    chan struct{}                       // notify net it should close
	StateGraphChannel chan string                         // used for writting the state graph to a file
	Timeout           time.Duration                       // maximum total time the net can run for
	Patience          time.Duration                       // maximum time between transitions
}

func (net *Net) NewNetFromJson(jsonNet NetJson) {
	net.NewNet()

	for _, place := range jsonNet.Places {
		net.AddPlace(place.Name, place.Tokens)
	}
	for _, transition := range jsonNet.Transitions {
		net.AddTransition(transition)
	}
	for _, edge := range jsonNet.Edges {
		net.AddEdge(edge.From, edge.To, edge.Weight)
	}
}

func (net *Net) NewNet() {
	net.Places = make(map[string]*Place)
	net.Transitions = make(map[string][]int)
	net.InEdges = make(map[string][]WeightedTransitionEdge)
	net.ReverseInEdges = make(map[string][]WeightedPlaceEdge)
	net.OutEdges = make(map[string][]WeightedPlaceEdge)
	net.ClosingChannel = make(chan struct{})
	net.StateGraphChannel = make(chan string, 100)
	net.Timeout = 5 * time.Second
	net.Patience = 100 * time.Millisecond
}

func (net *Net) WriteStateGraphToFile(initialPlaces map[string]int) {
	file_path, err := filepath.Abs("./data/state_graph.txt")
	PrintError(err)
	_ = os.Remove(file_path)
	f, err := os.OpenFile(file_path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	PrintError(err)
	_, err = f.WriteString("State Graph\n\n|")
	PrintError(err)
	// Initialize local places
	localPlaces := initialPlaces
	placeNames := make([]string, 0, len(localPlaces))
	for place := range localPlaces {
		placeNames = append(placeNames, place)
	}
	sort.Strings(placeNames)
	// Write the initial number of tokens in each place
	for _, placeName := range placeNames {
		_, err := f.WriteString(fmt.Sprintf(" %s: %d |", placeName, localPlaces[placeName]))
		PrintError(err)
	}
	// Write state graph to file
	for {
		select {
		case transition := <-net.StateGraphChannel:
			_, err := f.WriteString(fmt.Sprintf("\n\nFired transition %s\n\n|", transition))
			PrintError(err)
			// Modify the local tokens according to the transition
			for _, pe := range net.ReverseInEdges[transition] {
				localPlaces[pe.Place] -= pe.Weight
			}
			for _, pe := range net.OutEdges[transition] {
				localPlaces[pe.Place] += pe.Weight
			}
			// Write the number of tokens in each place
			for _, placeName := range placeNames {
				_, err := f.WriteString(fmt.Sprintf(" %s: %d |", placeName, localPlaces[placeName]))
				PrintError(err)
			}
		case <-time.After(net.Patience):
			net.ClosingChannel <- struct{}{}
			_, err = f.WriteString("\n")
			PrintError(err)
			f.Close()
			return
		}
	}
}

func (net *Net) AddPlace(place string, tokens int) {
	net.Places[place] = &Place{tokens, &sync.Mutex{}}
}

func (net *Net) AddTransition(transition string) {
	net.Transitions[transition] = []int{}
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
	initialPlaces := make(map[string]int)
	for placeName, place := range net.Places {
		initialPlaces[placeName] = place.tokens
	}
	for _, workCluster := range net.WorkClusters {
		workCluster.PrintWorkCluster()
		go workCluster.checkFire(net)
		workCluster.FireChannel <- struct{}{}
	}
	go net.TimeoutNet()
	go net.WriteStateGraphToFile(initialPlaces)
}

func (net *Net) TimeoutNet() {
	// Maximum time the net can run for
	<-time.After(net.Timeout)
	net.ClosingChannel <- struct{}{}
}

func (net *Net) CheckClosingChannel() {
	<-net.ClosingChannel

	for _, workCluster := range net.WorkClusters {
		close(workCluster.FireChannel)
	}

	fmt.Println("Closing Net")
}

func (net *Net) Fire(transition string) {
	// remove tokens from incoming places
	for _, pe := range net.ReverseInEdges[transition] {
		net.Places[pe.Place].mutex.Lock()
		net.Places[pe.Place].tokens -= pe.Weight
		net.Places[pe.Place].mutex.Unlock()
	}
	net.StateGraphChannel <- transition
	// add tokens to outgoing places
	for _, pe := range net.OutEdges[transition] {
		net.Places[pe.Place].mutex.Lock()
		net.Places[pe.Place].tokens += pe.Weight
		net.Places[pe.Place].mutex.Unlock()
	}
	for _, workCluster := range net.Transitions[transition] {
		select {
		case net.WorkClusters[workCluster].FireChannel <- struct{}{}: // send signal to work cluster if not empty
		default: // do nothing if channel is full
		}
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

func (net *Net) PrintTokens() {
	for place, p := range net.Places {
		fmt.Printf("\t%s: %d\n", place, p.tokens)
	}
}
