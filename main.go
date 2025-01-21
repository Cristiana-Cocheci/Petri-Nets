package main

import (
	"Petri-Nets/src"
	"fmt"
)

func main() {
	fmt.Println("Petri Nets for Concurrent Programming")

	netJson := src.ReadNetJson("data/net.json")

	net := src.Net{}
	net.NewNetFromJson(netJson)

	net.PrintTokens()

	net.Run()

	net.CheckClosingChannel()
	net.PrintTokens()
}
