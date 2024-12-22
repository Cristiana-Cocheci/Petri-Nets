package main

import (
	"Petri-Nets/src"
	"fmt"
)

func main() {

	fmt.Println("Petri Nets for Concurrent Programming")
	exampleNet := src.Net{}
	exampleNet.NewNet()
	exampleNet.AddPlace("p1", 1)
	exampleNet.AddPlace("p2", 0)
	exampleNet.AddPlace("p3", 0)

	exampleNet.AddTransition("t1")
	exampleNet.AddTransition("t2")

	exampleNet.AddEdge("p1", "t1", 1)
	exampleNet.AddEdge("p1", "t2", 3)
	exampleNet.AddEdge("t1", "p2", 1)
	exampleNet.AddEdge("p2", "t2", 1)
	exampleNet.AddEdge("t2", "p3", 1)
	exampleNet.AddEdge("p3", "t1", 1)

	exampleNet.PrintNet()

}
