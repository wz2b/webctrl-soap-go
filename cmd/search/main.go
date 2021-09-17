package main

import (
	"fmt"
	"os"
	alcsoap "webctrl-soap-go/pkg/webctrl_soap_go"
)

func main() {

	config, err := processArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid command line arguments: %s", err)
	}

	alc := alcsoap.NewSoapService(config.server, config.user, config.password)

	out := make(chan alcsoap.GqlNode)

	go getChildren(alc.Eval, config.start, func(node alcsoap.GqlNode) bool {
		return true
	}, out)

	for child := range out {
		fmt.Printf("%s %s \"%s\"\n", child.Type, child.ReferenceName, child.DisplayName)
	}
}

func getChildren(eval alcsoap.EvalService, gql string, filter func(alcsoap.GqlNode) bool, out chan<- alcsoap.GqlNode) {
	recurse(eval, gql, filter, out)
	close(out)
}

func recurse(eval alcsoap.EvalService, gql string, filter func(alcsoap.GqlNode) bool, out chan<- alcsoap.GqlNode) {
	children, err := eval.GetChildren(gql, filter)

	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to get children of %s: %s", gql, err)
		close(out)
	}

	for _, node := range children {
		node.ReferenceName = gql + "/" + node.ReferenceName
		out <- node
		switch node.Type {
		case "AREA", "EQUIPMENT":
			recurse(eval,
				node.ReferenceName, filter, out)
		}
	}
}
