package main

import (
	"fmt"

	"github.com/rezkam/papilot/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
