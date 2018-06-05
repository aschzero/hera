package main

import (
	"io/ioutil"
)

// CurrentVersion reads and returns the contents of the version file
func CurrentVersion() string {
	version, err := ioutil.ReadFile("/VERSION")
	if err != nil {
		log.Error(err)
	}

	return string(version)
}
