package main

import "testing"

func TestVerifyLabelConfig(t *testing.T) {
	labels := map[string]string{}
	c := &Container{
		ID:       "2134",
		Hostname: "host.name",
		Labels:   labels,
	}

	if err := c.VerifyLabelConfig(); err == nil {
		t.Error("want error for no labels")
	}

	labels["hera.hostname"] = "address"
	if err := c.VerifyLabelConfig(); err == nil {
		t.Error("want error for one hera label")
	}

	labels["hera.port"] = "80"
	if err := c.VerifyLabelConfig(); err != nil {
		t.Error(err)
	}
}
