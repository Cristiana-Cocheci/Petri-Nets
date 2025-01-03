package main

import (
	"Petri-Nets/src"
	"fmt"
	"time"
)

func main() {
	fmt.Println("Petri Nets for Concurrent Programming")

	netJson := src.ReadNetJson("net.json")

	net := src.Net{}
	net.NewNetFromJson(netJson)

	net.PrintNet()

	net.Run()
	<-time.After(3 * time.Second)
	net.PrintNet()

}
