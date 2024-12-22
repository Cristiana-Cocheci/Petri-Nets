package src

import "fmt"

func PrintError(err error) {
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
}
