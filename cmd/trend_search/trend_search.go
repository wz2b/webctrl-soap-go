package main

import (
	"fmt"
	"os"
)

func main() {

	config, err := processArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid command line arguments: %s", err)
	}

	fmt.Println(config.start)
}