package main

import (
	"fmt"
	"os"
	"time"
	alcsoap "webctrl-soap-go/pkg/webctrl_soap_go"
)

func main() {

	config, err := processArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid command line arguments: %s\n", err)
	}

	alc := alcsoap.NewSoapService(config.server, config.user, config.password)

	out := make(chan alcsoap.GqlNode)

	go getChildren(alc.Eval, config.start, func(node alcsoap.GqlNode) bool {
		return true || node.Type == "AREA" || node.Type == "EQUIPMENT"
	}, out)

	types := make(map[string]int)

	for child := range out {
		fmt.Printf("%s %s \"%s\"\n", child.Type, child.ReferenceName, child.DisplayName)
		t, ok := types[child.Type]
		if ok {
			types[child.Type] = t + 1
		} else {
			types[child.Type] = 1
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
