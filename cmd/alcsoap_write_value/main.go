package main

import (
	"fmt"
	"os"
	alcsoap "webctrl-soap-go/pkg/webctrl_soap_go"
)

func main() {
	config, err := processArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid command line arguments: %s\n", err)
		os.Exit(1)
	}

	alc := alcsoap.NewSoapService(config.server, config.user, config.password)

	if config.verbose {
		fmt.Fprintf(os.Stderr, "Writing value to: %s\n", config.gql)
		fmt.Fprintf(os.Stderr, "Value: %s\n", config.value)
		fmt.Fprintf(os.Stderr, "Reason: %s\n", config.reason)
	}

	err = alc.Eval.SetValue(config.gql, config.value, config.reason)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing value: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully wrote value to %s\n", config.gql)
}
