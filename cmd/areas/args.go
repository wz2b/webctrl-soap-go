package main

import (
	"flag"
	"fmt"
)

func processArgs() (config CommandLineOpts, err error) {
	flag.StringVar(
		&config.credentialsFile,
		"creds",
		"",
		"WebCTRL credentials file - file must be one line containing user:password",
	)
	flag.StringVar(
		&config.user,
		"user",
		"",
		"WebCTRL username (must have SOAP privileges)",
	)
	flag.StringVar(
		&config.password,
		"password",
		"",
		"WebCTRL password",
	)
	flag.StringVar(
		&config.start,
		"start",
		"/trees/geographic",
		"GQL string of start location",
	)
	flag.StringVar(
		&config.pattern,
		"search",
		"",
		"Search pattern",
	)
	flag.BoolVar(
		&config.verbose,
		"v",
		false,
		"Display extra information",
	)
	flag.StringVar(
		&config.server,
		"server",
		"",
		"URL to WebCTRL server",
	)

	if !flag.Parsed() {
		flag.Parse()
	}

	if config.server == "" {
		return config, fmt.Errorf("must specify a server")
	}

	if config.credentialsFile != "" {
		usernameFromFile, passwordFromFile, loadCredentialsErr :=
			loadCredentialsFromFile(config.credentialsFile)

		if loadCredentialsErr != nil {
			return config, loadCredentialsErr
		}

		if config.user == "" {
			config.user = usernameFromFile
		}

		if config.password == "" {
			config.password = passwordFromFile
		}
	}

	if config.user == "" {
		return config, fmt.Errorf(
			"username is required either on the command line or credentials file",
		)
	}

	if config.password == "" {
		return config, fmt.Errorf(
			"password is required either on the command line or credentials file",
		)
	}

	if len(flag.Args()) != 0 {
		return config, fmt.Errorf(
			"unexpected positional arguments: %v",
			flag.Args(),
		)
	}

	return config, nil
}
