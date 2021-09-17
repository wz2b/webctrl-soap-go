package main

import (
	"errors"
	"io/ioutil"
	"strings"
)

func loadCredentialsFromFile(path string) (username string, password string, err error) {
	data, err := ioutil.ReadFile(path)

	if err != nil {
		str := string(data)
		str = strings.TrimSpace(str)
		fields := strings.Split(str, ":")
		if len(fields) != 2 {
			err = errors.New("credentials file is invalid  Must be one line of username:password")
		} else {
			username = fields[0]
			password = fields[1]
		}
	}
	return
}
