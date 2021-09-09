package main

import (
	"errors"
	"flag"
	"github.com/Songmu/prompter"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"time"
)

func processArgs() (config CommandLineOpts, err error) {
	flag.StringVar(&config.credentialsFile, "creds", "", "WebCTRL credentials file - file must be one line containing user:password")
	flag.StringVar(&config.user, "user", "", "WebCTRL username (must have SOAP privileges)")
	flag.StringVar(&config.password, "password", "", "WebCTRL password")
	flag.StringVar(&config.startStr, "start", "", "Start time (inclusive)")
	flag.StringVar(&config.stopStr, "stop", "", "End time (exclusive)")
	flag.BoolVar(&config.verbose, "v", false, "Display extra information")
	flag.StringVar(&config.configFilePath, "config", "", "YAML configuration file")
	cmdLineServer := flag.String("server", "", "URL to WebCTRL server")

	if !flag.Parsed() {
		flag.Parse()
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

	if config.startStr == "" {
		config.start = time.Now().Add(-24 * time.Hour)
	} else {
		config.start, err = time.ParseInLocation(LAYOUT, config.startStr, time.Local)
		if err != nil {
			err = errors.New("could not parse start time")
			return
		}
	}

	if config.stopStr == "" {
		// If the user doesn't specify a time, use 1 second from now - that will get
		// everything up to and including now (if such a sample is available).
		config.stop = time.Now().Add(1 * time.Second)
	} else {
		config.stop, err = time.ParseInLocation(LAYOUT, config.stopStr, time.Local)
		if err != nil {
			err = errors.New("could not parse stop time")
			return
		}
	}

	if config.configFilePath != "" {
		file, readFileError := ioutil.ReadFile(config.configFilePath)
		if readFileError != nil {
			err = readFileError
			return
		}
		err = yaml.Unmarshal(file, &config.ConfigFile)
	}

	// The remainder of the command line is a list of locations.  To accept these,
	// the user must have specified a server.  If present this server and the additional
	// locations will be added to the end of the config file specified location list.
	cmdLineLocations := flag.Args()
	if len(cmdLineLocations) > 0 && (cmdLineServer == nil || len(*cmdLineServer) == 0) {
		err = errors.New("must include a -server option if specifying locations from the command line")
		return
	}

	if cmdLineLocations != nil && len(cmdLineLocations) > 0 {
		flagServer := ServerConfig{Url: *cmdLineServer, Login: config.user, Password: config.password}
		flagLocations := make([]TrendLocationConfig, len(cmdLineLocations))
		for i, loc := range cmdLineLocations {
			flagLocations[i] = TrendLocationConfig{Location: loc}
		}
		flagServer.Locations = flagLocations
		config.ConfigFile.Servers = append(config.ConfigFile.Servers, flagServer)
	}

	if len(cmdLineLocations) > 0 && config.user == "" {
		// No username specified either in a credentials file or on the command line.
		// Try prompting for one.
		config.user = prompter.Prompt("WebCTRL username", "")
		if config.user == "" {
			// Still no username - error out
			err = errors.New("username must be specified either on the command line or in a credentials file")
			return
		}
	}

	if len(cmdLineLocations) > 0 && config.password == "" {
		// No password specified either in a credentials file or on the command line.
		// Try prompting for one.
		config.password = prompter.Password("WebCTRL password")
		if config.password == "" {
			// Still no password - error out
			err = errors.New("password must be specified either on the command line or in a credentials file")
			return
		}
	}

	//
	// Any servers specified in the config file that don't have their own
	// login or password specified will get the one provided on the
	// command line / credentials file
	//
	for i, server := range config.ConfigFile.Servers {
		if server.Login == "" && config.user != "" {
			config.ConfigFile.Servers[i].Login = config.user
		}

		if server.Password == "" && config.password != "" {
			config.ConfigFile.Servers[i].Password = config.password
		}
	}

	return
}
