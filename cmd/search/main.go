package main

import (
	"fmt"
	"os"
	"time"
	alcsoap "github.com/wz2b/webctrl-soap-go"
)

func main() {

	config, err := processArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid command line arguments: %s\n", err)
		os.Exit(1)
	}

	alc := alcsoap.NewSoapService(config.server, config.user, config.password)

	out := make(chan alcsoap.GqlNode)

	fmt.Printf("Starting search at %s\n", config.start)
	go getChildren(alc.Eval, config.start, func(node alcsoap.GqlNode) bool {
		return true
	}, out)

	types := make(map[string]int)
	i := 0

	for child := range out {
		fmt.Printf("%s %s \"%s\"\n", child.Type, child.ReferenceName, child.DisplayName)
		t, ok := types[child.Type]
		if ok {
			types[child.Type] = t + 1
		} else {
			types[child.Type] = 1
		}

		if i < 50 {
			i = i + 1
		} else {
			i = 0

			fmt.Println("\nTypes so far\n============")
			for k, v := range types {
				fmt.Printf("%s\t%d\n", k, v)
			}

		}

	}

	fmt.Println("\nFinal count of types")
	for k, v := range types {
		fmt.Printf("%s\t%d\n", k, v)
	}

}

func getChildren(eval alcsoap.EvalService, gql string, filter func(alcsoap.GqlNode) bool, out chan<- alcsoap.GqlNode) {
	recurse(eval, gql, filter, out)
	close(out)
}

func recurse(eval alcsoap.EvalService, gql string, filter func(alcsoap.GqlNode) bool, out chan<- alcsoap.GqlNode) {

	time.Sleep(1 * time.Second)

	children, err := eval.GetChildren(gql, filter)

	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to get children of %s: %s", gql, err)
		close(out)
	}

	for _, node := range children {
		node.ReferenceName = gql + "/" + node.ReferenceName
		out <- node
		switch node.Type {
		case "AREA",
			"EQUIPMENT", "BEQU",
			"BAI", "BAO", "BAV",
			"BBI", "BBO", "BBV",
			"LPOINT", // ANI2, BNI2, etc.
			"BMAI", "BMBO", "BMBV",
			"BMSI", "BMSO", "BMSV":
			recurse(eval, node.ReferenceName, filter, out)
		}
	}
}
