package main

// CommandLineOpts holds command line arguments
type CommandLineOpts struct {
	server          string
	user            string
	password        string
	credentialsFile string
	verbose         bool
	start           string
	pattern         string
}
