package main

// CommandLineOpts holds command line arguments
type CommandLineOpts struct {
	server          string
	user            string
	password        string
	credentialsFile string
	verbose         bool
	gql             string
	value           string
	reason          string
}
