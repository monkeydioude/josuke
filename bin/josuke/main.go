package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/monkeydioude/josuke"
)

// findOutProtocolHandler defines which handler, with respect to the protocol,
// should be used, from a josuke.Josuke struct
func findOutProtocolHandler(j *josuke.Josuke) (string, func() error) {
	p := fmt.Sprintf("%s:%d", j.Host, j.Port)
	if j.Key == "" {
		return "http", func() error {
			return http.ListenAndServe(p, nil)
		}
	}

	return "https", func() error {
		return http.ListenAndServeTLS(p, j.Cert, j.Key, nil)
	}
}

func main() {
	configFileName := flag.String("c", "config.json", "Path to config file, yml or json format")
	flag.Parse()

	j, err := josuke.New(*configFileName)

	if err != nil {
		log.Fatal("[ERR ] ", err)
	}

	j.HandleHooks()

	if j.HealthcheckRoute == "" {
		j.HealthcheckRoute = "/healthcheck"
	}
	http.HandleFunc(j.HealthcheckRoute, HealthcheckHandler)

	protocol, handler := findOutProtocolHandler(j)
	if j.LogEnabled(josuke.InfoLevel) {
		log.Printf("[INFO] Listening %s://%s:%d\n", protocol, j.Host, j.Port)
	}
	log.Fatal(handler())
}
