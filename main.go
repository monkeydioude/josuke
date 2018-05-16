package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func getCwd() string {
	ex, err := os.Executable()

	if err != nil {
		log.Fatal("Could not resolve os.Executable")
	}

	return filepath.Dir(ex)
}

func main() {
	configFileName := flag.String("c", "config.json", "Path to config file")
	port := flag.Int("p", 8082, "Port server will listen to")
	uri := flag.String("u", "josuke", "URI webhook will listen to")
	flag.Parse()

	file, err := ioutil.ReadFile(*configFileName)

	if err != nil {
		log.Fatalf("Could not read config file: %v", err)
	}

	if err := json.Unmarshal(file, &Config); err != nil {
		log.Fatalf("Could not parse json from config file")
	}

	http.HandleFunc(fmt.Sprintf("/%s/github", *uri), GithubRequest)
	http.HandleFunc(fmt.Sprintf("/%s/bitbucket", *uri), BitbucketRequest)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
