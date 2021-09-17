package main

import (
	"flag"
	"fmt"
)

func processArgs() (config CommandLineOpts, err error) {
	flag.StringVar(&config.credentialsFile, "creds", "", "WebCTRL credentials file - file must be one line containing user:password")
	flag.StringVar(&config.user, "user", "", "WebCTRL username (must have SOAP privileges)")
	flag.StringVar(&config.password, "password", "", "WebCTRL password")
	flag.StringVar(&config.start, "start", "", "GQL string of start location")
	flag.StringVar(&config.pattern, "search", "", "Start time (inclusive)")
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

	// fetch the rest of the command line
	cmdLineLocations := flag.Args()

	// Validation
	if config.user == "" {
		return config, fmt.Errorf("username is required either on the command line or credentials file")
	}

	if config.password == "" {
		return config, fmt.Errorf("password is required either on the command line or credentials file")
	}

	switch len(cmdLineLocations) {
	case 0:
		config.start = "/trees/geographic"
	case 1:
		config.start = cmdLineLocations[0]

	default:
		return config, fmt.Errorf("must specify a starting location")
	}

	return config, nil
}
