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

	alc := alcsoap.NewSoapService(
		config.server,
		config.user,
		config.password,
	)

	out := make(chan alcsoap.GqlNode)

	go getChildren(
		alc.Eval,
		config.start,
		func(node alcsoap.GqlNode) bool {
			// Discovery mode: accept every node type.
			return true
		},
		out,
	)

	types := make(map[string]int)

	for child := range out {
		fmt.Printf(
			"%s %s %q\n",
			child.Type,
			child.ReferenceName,
			child.DisplayName,
		)

		types[child.Type]++
	}

	fmt.Println("\nFinal count of types")

	for k, v := range types {
		fmt.Printf("%s\t%d\n", k, v)
	}
}

func getChildren(
	eval alcsoap.EvalService,
	gql string,
	filter func(alcsoap.GqlNode) bool,
	out chan<- alcsoap.GqlNode,
) {
	defer close(out)

	recurse(eval, gql, filter, out)
}

func recurse(
	eval alcsoap.EvalService,
	gql string,
	filter func(alcsoap.GqlNode) bool,
	out chan<- alcsoap.GqlNode,
) {
	time.Sleep(1 * time.Second)

	fmt.Fprintf(os.Stderr, "GET %s\n", gql)

	children, err := eval.GetChildren(gql, filter)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"unable to get children of %s: %s\n",
			gql,
			err,
		)
		return
	}

	fmt.Fprintf(
		os.Stderr,
		"    %d children\n",
		len(children),
	)

	for _, node := range children {
		node.ReferenceName = gql + "/" + node.ReferenceName

		// Emit every node regardless of type.
		out <- node

		// Discovery mode: attempt to recurse into every node.
		//
		// Leaf nodes should simply return zero children. This lets us
		// discover container types that we don't know about yet.
		recurse(
			eval,
			node.ReferenceName,
			filter,
			out,
		)
	}
}
