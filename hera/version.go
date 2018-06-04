package main

import (
	"io/ioutil"
)

func CurrentVersion() string {
	version, err := ioutil.ReadFile("/VERSION")
	if err != nil {
		log.Error(err)
	}

	return string(version)
}
