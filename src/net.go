package src

import "fmt"

type WeightedTransitionEdge struct {
	Transition string
	Weight     int
}

func (edge *WeightedTransitionEdge) PrintEdge() {
	fmt.Printf("		Transition %s: Weight %d\n", edge.Transition, edge.Weight)
}

type WeightedPlaceEdge struct {
	Place  string
	Weight int
}

func (edge *WeightedPlaceEdge) PrintEdge() {
	fmt.Printf("		Place %s: Weight %d\n", edge.Place, edge.Weight)
}

type Net struct {
	Places      map[string]int // holds number of current tokens
	Transitions map[string]struct{}
	InEdges     map[string][]WeightedTransitionEdge // places -> transitions
	OutEdges    map[string][]WeightedPlaceEdge      // transitions -> places
}

func (net *Net) NewNet() {
	net.Places = make(map[string]int)
	net.Transitions = make(map[string]struct{})
	net.InEdges = make(map[string][]WeightedTransitionEdge)
	net.OutEdges = make(map[string][]WeightedPlaceEdge)
}

func (net *Net) AddPlace(place string, tokens int) {
	net.Places[place] = tokens
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
	} else if fromIsTransition && toIsPlace {
		net.OutEdges[from] = append(net.OutEdges[from], WeightedPlaceEdge{Place: to, Weight: weight})
	} else {
		PrintError(fmt.Errorf("invalid edge from %s to %s", from, to))
	}
}

func (net *Net) PrintNet() {
	fmt.Println("Places:")
	for place, tokens := range net.Places {
		fmt.Printf("	%s: %d\n", place, tokens)
	}
	fmt.Println("Transitions:")
	for transition := range net.Transitions {
		fmt.Printf("	%s\n", transition)
	}
	fmt.Println("Edges:")
	for place, edges := range net.InEdges {
		fmt.Printf("	Place %s -> \n", place)
		for _, edge := range edges {
			edge.PrintEdge()
		}
	}
	for transition, edges := range net.OutEdges {
		fmt.Printf("	Transition %s -> \n", transition)
		for _, edge := range edges {
			edge.PrintEdge()
		}
	}
}
