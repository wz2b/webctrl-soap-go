// args.go
package main

import (
	"flag"
	"fmt"
	"strings"
)

func processArgs() (config CommandLineOpts, err error) {
	flag.StringVar(&config.credentialsFile, "creds", "", "WebCTRL credentials file - file must be one line containing user:password")
	flag.StringVar(&config.user, "user", "", "WebCTRL username (must have SOAP privileges)")
	flag.StringVar(&config.password, "password", "", "WebCTRL password")
	flag.StringVar(&config.gql, "gql", "", "GQL string of the object to write value to")
	flag.StringVar(&config.value, "value", "", "Value to write (required)")
	flag.StringVar(&config.reason, "reason", "CLI Write", "Change reason to log with the write")
	flag.BoolVar(&config.verbose, "v", false, "Display extra information")
	flag.StringVar(&config.server, "server", "", "URL to WebCTRL server")

	if !flag.Parsed() {
		flag.Parse()
	}

	if config.server == "" {
		err = fmt.Errorf("Must specify a server")
		return config, err
	}

	if config.credentialsFile != "" {
		usernameFromFile, passwordFromFile, loadCredentialsErr := loadCredentialsFromFile(config.credentialsFile)
		if loadCredentialsErr != nil {
			err = loadCredentialsErr
			return config, err
		}

		if config.user == "" {
			config.user = usernameFromFile
		}

		if config.password == "" {
			config.password = passwordFromFile
		}
	}

	// Validation
	if config.user == "" {
		return config, fmt.Errorf("username is required either on the command line or credentials file")
	}

	if config.password == "" {
		return config, fmt.Errorf("password is required either on the command line or credentials file")
	}

	if config.gql == "" {
		return config, fmt.Errorf("--gql is required (GQL path to the object)")
	}

	if strings.ToLower(config.value) == "null" {
		config.value = ""
	} else {
		if config.value == "" {
			return config, fmt.Errorf("--value is required (value to write)")
		}
	}

	return config, nil
}
