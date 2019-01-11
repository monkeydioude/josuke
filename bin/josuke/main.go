package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/monkeydioude/josuke"
)

func getCwd() string {
	ex, err := os.Executable()

	if err != nil {
		log.Fatal("[ERR ] Could not resolve os.Executable")
	}

	return filepath.Dir(ex)
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

	p := fmt.Sprintf(":%d", j.Port)
	log.Printf("[INFO] Listening port %s\n", p)
	log.Fatal(http.ListenAndServe(p, nil))
}
