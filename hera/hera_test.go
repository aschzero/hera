package main

import (
	"testing"
)

func TestVerifyLabelConfig(t *testing.T) {
	labels := map[string]string{}

	if err := verifyLabelConfig(labels); err == nil {
		t.Error(err)
	}

	labels["hera.hostname"] = "address"
	if err := verifyLabelConfig(labels); err == nil {
		t.Error(err)
	}

	labels["hera.port"] = "80"
	if err := verifyLabelConfig(labels); err != nil {
		t.Error(err)
	}
}
