package main

import (
	"errors"
	"flag"
	"time"
)

type ConfigType struct {
	user     string
	password string
	server   string
	startStr string
	stopStr  string
	verbose  bool
	start    time.Time
	stop     time.Time
	args     []string
}

func processArgs() (ConfigType, error) {
	var config ConfigType

	flag.StringVar(&config.user, "user", "", "WebCTRL username (must have SOAP privileges)")
	flag.StringVar(&config.password, "password", "", "WebCTRL password")
	flag.StringVar(&config.server, "server", "", "URL to WebCTRL server")
	flag.StringVar(&config.startStr, "start", "", "Start time (inclusive)")
	flag.StringVar(&config.stopStr, "stop", "", "End time (exclusive)")
	flag.BoolVar(&config.verbose, "v", false, "Display extra information")

	// TODO: make a way to specify the list of trends from a file
	// TODO: make a way to specify the server credentials from a file

	if !flag.Parsed() {
		flag.Parse()
		config.args = flag.Args()
	}
	var err error

	if config.user == "" {
		return config, errors.New("user is a required argument")
	}

	if config.password == "" {
		return config, errors.New("password is a required argument")
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

	config.args = flag.Args()
	return config, nil
}
