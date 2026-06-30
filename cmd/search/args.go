package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

func processArgs() (config CommandLineOpts, err error) {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&config.credentialsFile, "creds", "", "WebCTRL credentials file - file must be one line containing user:password")
	fs.StringVar(&config.user, "user", "", "WebCTRL username (must have SOAP privileges)")
	fs.StringVar(&config.password, "password", "", "WebCTRL password")
	fs.StringVar(&config.start, "start", "", "GQL string of start location")
	fs.StringVar(&config.pattern, "search", "", "Start time (inclusive)")
	fs.BoolVar(&config.verbose, "v", false, "Display extra information")
	fs.StringVar(&config.server, "server", "", "URL to WebCTRL server")

	if err = fs.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return config, err
		}
		if err.Error() == "flag needs an argument: -start" {
			return config, fmt.Errorf("%w; if the start GQL begins with #, pass it as -start='#...' or -start=#...", err)
		}
		return config, err
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
	cmdLineLocations := fs.Args()

	// Validation
	if config.user == "" {
		return config, fmt.Errorf("username is required either on the command line or credentials file")
	}

	if config.password == "" {
		return config, fmt.Errorf("password is required either on the command line or credentials file")
	}

	switch len(cmdLineLocations) {
	case 0:
		if config.start == "" {
			config.start = "/trees/geographic"
		}
	case 1:
		if config.start == "" {
			config.start = cmdLineLocations[0]
		}
	default:
		return config, fmt.Errorf("must specify a starting location")
	}

	if config.start == "" {
		return config, fmt.Errorf("must specify a starting location")
	}

	return config, nil
}
