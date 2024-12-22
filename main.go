package main

import (
	"Petri-Nets/src"
	"fmt"
)

func main() {
	fmt.Println("Petri Nets for Concurrent Programming")

	exampleNet := src.Net{}
	exampleNet.NewNet()
	// Place nodes
	exampleNet.AddPlace("p1", 1)
	exampleNet.AddPlace("p2", 0)
	exampleNet.AddPlace("p3", 0)
	exampleNet.AddPlace("p4", 5)
	// Transition nodes
	exampleNet.AddTransition("t1")
	exampleNet.AddTransition("t2")
	exampleNet.AddTransition("t3")
	// Place -> Transition edges
	exampleNet.AddEdge("p1", "t1", 1)
	exampleNet.AddEdge("p1", "t2", 3)
	exampleNet.AddEdge("p2", "t2", 1)
	exampleNet.AddEdge("p3", "t1", 1)
	exampleNet.AddEdge("p4", "t3", 2)
	// Transition -> Place edges
	exampleNet.AddEdge("t1", "p2", 1)
	exampleNet.AddEdge("t2", "p3", 1)
	exampleNet.AddEdge("t3", "p2", 1)
	exampleNet.AddEdge("t3", "p3", 0)

	exampleNet.PrintNet()

	workClusters := exampleNet.SplitNet()
	for _, workCluster := range workClusters {
		workCluster.PrintWorkCluster()
	}
}
