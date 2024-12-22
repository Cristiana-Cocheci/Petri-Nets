package src

import (
	"fmt"
	"sync"
)

type WeightedTransitionEdge struct {
	Transition string
	Weight     int
}

func (edge *WeightedTransitionEdge) PrintEdge() {
	fmt.Printf("\t\tTransition %s: Weight %d\n", edge.Transition, edge.Weight)
}

type WeightedPlaceEdge struct {
	Place  string
	Weight int
}

func (edge *WeightedPlaceEdge) PrintEdge() {
	fmt.Printf("\t\tPlace %s: Weight %d\n", edge.Place, edge.Weight)
}

type Place struct {
	tokens int
	mutex  *sync.Mutex
}

type Net struct {
	Places         map[string]*Place // holds number of current tokens
	Transitions    map[string]struct{}
	InEdges        map[string][]WeightedTransitionEdge // places -> transitions
	ReverseInEdges map[string][]WeightedPlaceEdge
	OutEdges       map[string][]WeightedPlaceEdge // transitions -> places
}

func (net *Net) NewNet() {
	net.Places = make(map[string]*Place)
	net.Transitions = make(map[string]struct{})
	net.InEdges = make(map[string][]WeightedTransitionEdge)
	net.ReverseInEdges = make(map[string][]WeightedPlaceEdge)
	net.OutEdges = make(map[string][]WeightedPlaceEdge)
	go net.checkFire()
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

func (net *Net) checkFire() {
	for {
		// check if transition can be fired (enough tokens collected in incoming places)
		canFire := true
		for transition, placeEdge := range net.ReverseInEdges {
			for _, pe := range placeEdge {
				weight := pe.Weight
				incomingPlaces := net.Places[pe.Place]
				if incomingPlaces.tokens < weight {
					canFire = false
					break
				}
			}
			if canFire {
				net.Fire(transition)
			}

		}
	}
}

func (net *Net) Fire(transition string) {
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
