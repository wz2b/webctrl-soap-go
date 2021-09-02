package main

import (
	"errors"
	"flag"
	"github.com/Songmu/prompter"
	"time"
)

type ConfigType struct {
	user            string
	password        string
	credentialsFile string
	configFile      string
	server          string
	startStr        string
	stopStr         string
	verbose         bool
	start           time.Time
	stop            time.Time
	args            []string
}

func processArgs() (config ConfigType, err error) {

	flag.StringVar(&config.credentialsFile, "creds", "", "WebCTRL credentials file - file must be one line containing user:password")
	flag.StringVar(&config.user, "user", "", "WebCTRL username (must have SOAP privileges)")
	flag.StringVar(&config.password, "password", "", "WebCTRL password")
	flag.StringVar(&config.server, "server", "", "URL to WebCTRL server")
	flag.StringVar(&config.startStr, "start", "", "Start time (inclusive)")
	flag.StringVar(&config.stopStr, "stop", "", "End time (exclusive)")
	flag.BoolVar(&config.verbose, "v", false, "Display extra information")
	flag.StringVar(&config.configFile, "config", "", "YAML configuration file")

	// TODO: make a way to specify the list of trends from a file

	if !flag.Parsed() {
		flag.Parse()
		config.args = flag.Args()
	}

	//
	// If a credentials file is specified, load it but user or password on the command
	// line will override
	//
	if config.credentialsFile != "" {
		usernameFromFile, passwordFromFile, loadCredentialsErr := loadCredentialsFromFile(config.credentialsFile)
		if loadCredentialsErr != nil {
			err = loadCredentialsErr
			return
		}

		if config.user == "" {
			config.user = usernameFromFile
		}

		if config.password == "" {
			config.password = passwordFromFile
		}
	}

	if config.user == "" {
		// No username specified either in a credentials file or on the command line.
		// Try prompting for one.
		config.user = prompter.Prompt("WebCTRL username", "")
		if config.user == "" {
			// Still no username - error out
			err = errors.New("username must be specified either on the command line or in a credentials file")
			return
		}
	}

	if config.password == "" {
		// No password specified either in a credentials file or on the command line.
		// Try prompting for one.
		config.password = prompter.Password("WebCTRL password")
		if config.password == "" {
			// Still no password - error out
			err = errors.New("password must be specified either on the command line or in a credentials file")
			return
		}
	}

	if config.server == "" {
		return config, errors.New("server is a required argument")
	}

	if config.startStr == "" {
		config.start = time.Now().Add(-24 * time.Hour)
	} else {
		config.start, err = time.ParseInLocation(LAYOUT, config.startStr, time.Local)
		if err != nil {
			return config, errors.New("could not parse start time")
		}
	}

	if config.stopStr == "" {
		// If the user doesn't specify a time, use 1 second from now - that will get
		// everything up to and including now (if such a sample is available).
		config.stop = time.Now().Add(1 * time.Second)
	} else {
		config.stop, err = time.ParseInLocation(LAYOUT, config.stopStr, time.Local)
		if err != nil {
			return config, errors.New("could not parse stop time")
		}
	}

	// Capture the remaining command line arguments
	config.args = flag.Args()
	return
}
