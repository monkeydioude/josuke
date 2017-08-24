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

	josuke "github.com/monkeydioude/josuke"
)

func getCwd() string {
	ex, err := os.Executable()

	if err != nil {
		log.Fatal("Could not resolve os.Executable")
	}

	return filepath.Dir(ex)
}

func main() {
	configFileName := fmt.Sprintf("%s/%s", getCwd(), *flag.String("c", "config.json", "Path to config file"))
	port := flag.Int("p", 8082, "Port server will listen to")
	uri := flag.String("u", "", "URI webhook will listen to")
	flag.Parse()

	file, err := ioutil.ReadFile(configFileName)

	if err != nil {
		log.Fatalf("Could not read config file: %v", err)
	}

	if err := json.Unmarshal(file, &josuke.Config); err != nil {
		log.Fatalf("Could not parse json from config file")
	}

	http.HandleFunc(fmt.Sprintf("/%s", *uri), josuke.Request)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
