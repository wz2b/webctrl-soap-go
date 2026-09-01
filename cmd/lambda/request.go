package main

import "encoding/json"

type LambdaRequest struct {
	JobName         string         `json:"jobName"`
	Server          string         `json:"server"`
	Username        string         `json:"username"`
	PasswordSsmPath string         `json:"passwordSsmPath"`
	RequestedBy     string         `json:"requestedBy,omitempty"`
	Writes          []WriteRequest `json:"writes"`
}

type WriteRequest struct {
	Path  string          `json:"path"`
	Value json.RawMessage `json:"value"`
}
