package main

import (
	"net/http"
	"testing"

	"github.com/monkeydioude/josuke"
)

func TestIShouldGiveHTTPOnEmptyKey(t *testing.T) {
	if p, _ := findOutProtocolHandler(&josuke.Josuke{
		Key: "",
	}); p != "http" {
		t.Fail()
	}
}

func TestIShouldReturnHTTPSProtocolWhenKeyExists(t *testing.T) {
	if p, _ := findOutProtocolHandler(&josuke.Josuke{
		Key: "Alicia",
	}); p != "https" {
		t.Fail()
	}
}

func TestIShouldThrowErrorWithHTPPSOnMissingCert(t *testing.T) {
	j := &josuke.Josuke{
		Key: "t",
	}
	if p, _ := findOutProtocolHandler(j); p != "https" {
		t.Fail()
	}
	e := http.ListenAndServeTLS("localhost:8082", j.Cert, j.Key, nil)
	if e == nil {
		t.Errorf("error should be triggerd")
	}
}
