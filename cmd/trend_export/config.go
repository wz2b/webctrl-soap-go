package main

import "time"

// CommandLineOpts holds command line arguments
type CommandLineOpts struct {
	user            string
	password        string
	credentialsFile string
	configFilePath  string
	startStr        string
	stopStr         string
	verbose         bool
	start           time.Time
	stop            time.Time
	ConfigFile      ConfigOpts
}

// ConfigOpts holds configuration info from a file
type ConfigOpts struct {
	Servers []ServerConfig
}

type ServerConfig struct {
	Url       string
	Login     string
	Password  string
	Locations []TrendLocationConfig
}

type TrendLocationConfig struct {
	Location      string
	Name          string
	Interpolation string
}
