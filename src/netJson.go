package src

import (
	"encoding/json"
	"os"
)

type PlaceJson struct {
	Name   string
	Tokens int
}

type EdgeJson struct {
	From   string
	To     string
	Weight int
}

type NetJson struct {
	Places      []PlaceJson
	Transitions []string
	Edges       []EdgeJson
}

func ReadFile(path string) []byte {
	fstr, err := os.ReadFile(path)
	PrintError(err)
	return fstr
}

func ReadNetJson(path string) NetJson {
	data := ReadFile(path)
	var NetJson NetJson

	err := json.Unmarshal(data, &NetJson)
	PrintError(err)
	return NetJson
}
