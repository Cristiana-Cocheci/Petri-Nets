package main

import (
	"Petri-Nets/src"
	"fmt"
	"time"
)

func main() {
	fmt.Println("Petri Nets for Concurrent Programming")

	net := src.Net{}
	net.NewNet()
	// Place nodes
	net.AddPlace("p1", 10)
	net.AddPlace("p2", 12)
	net.AddPlace("p3", 34)
	net.AddPlace("p4", 0)
	net.AddPlace("p5", 0)
	net.AddPlace("p6", 0)
	net.AddPlace("p7", 0)
	// Transition nodes
	net.AddTransition("t1")
	net.AddTransition("t2")
	net.AddTransition("t3")
	net.AddTransition("t4")
	net.AddTransition("t5")
	// Place -> Transition edges
	net.AddEdge("p1", "t1", 3)
	net.AddEdge("p2", "t1", 2)
	net.AddEdge("p3", "t2", 10)
	net.AddEdge("p4", "t3", 2)
	net.AddEdge("p4", "t4", 1)
	net.AddEdge("p5", "t4", 1)
	net.AddEdge("p6", "t5", 3)
	// Transition -> Place edges
	net.AddEdge("t1", "p4", 1)
	net.AddEdge("t2", "p4", 1)
	net.AddEdge("t2", "p5", 2)
	net.AddEdge("t4", "p7", 2)
	net.AddEdge("t3", "p6", 3)
	net.AddEdge("t5", "p1", 3)

	net.PrintNet()

	net.Run()
	<-time.After(3 * time.Second)
	net.PrintNet()

}
