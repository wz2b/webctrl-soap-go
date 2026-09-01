package main

type LambdaResponse struct {
	JobName         string        `json:"jobName"`
	RequestedBy     string        `json:"requestedBy,omitempty"`
	WritesRequested int           `json:"writesRequested"`
	WritesSucceeded int           `json:"writesSucceeded"`
	WritesFailed    int           `json:"writesFailed"`
	Results         []WriteResult `json:"results"`
}

type WriteResult struct {
	Path   string `json:"path"`
	Value  any    `json:"value,omitempty"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}
