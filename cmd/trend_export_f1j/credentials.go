package main

import (
	"errors"
	"io/ioutil"
	"strings"
)

func loadCredentialsFromFile(path string) (username string, password string, err error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", "", err
	}

	str := strings.TrimSpace(string(data))
	fields := strings.SplitN(str, ":", 2)

	if len(fields) != 2 {
		return "", "", errors.New(
			"credentials file is invalid; must be one line of username:password",
		)
	}

	username = fields[0]
	password = fields[1]

	return
}
