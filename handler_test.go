package main

import (
	"testing"
)

func TestGetRootDomain(t *testing.T) {
	domains := map[string]string{
		"sub.domain.com": "domain.com",
		"domain.net.za": "domain.net.za",
		"sub.domain.org.au": "domain.org.au",
	}

	for domain, expected := range domains {
		actual, err := getRootDomain(domain)

		if err != nil {
			t.Errorf("Got error: %v", err)
		}

		if actual != expected {
			t.Errorf("Unexpected domain, got %s", actual)
		}
	}
}