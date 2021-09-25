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
	configFileName := flag.String("c", "config.json", "Path to config file")
	flag.Parse()

	j, err := josuke.New(*configFileName)

	if err != nil {
		log.Printf("[ERR ] %s", err)
	}

	if j.BitbucketHook == "" && j.GithubHook == "" {
		log.Println("[ERR ] MUDA MUDA MUDA ! Josuke needs to handle at least one type of hook. See README.md for help")
	}

	if j.GithubHook != "" {
		http.HandleFunc(j.GithubHook, j.GithubRequest)
		log.Println("[INFO] Gureto daze 8), handling Github hooks")
	}
	if j.BitbucketHook != "" {
		http.HandleFunc(j.BitbucketHook, j.BitbucketRequest)
		log.Println("[INFO] Gureto daze 8), handling Bitbucket hooks")
	}

	protocol, handler := findOutProtocolHandler(j)
	log.Printf("[INFO] Listening %s://%s:%d\n", protocol, j.Host, j.Port)
	log.Fatal(handler())
}
