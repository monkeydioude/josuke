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
	"time"

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
	uri := flag.String("u", "josuke", "URI webhook will listen to")
	httpPort := flag.Int("hp", 8083, "Http Server Listener")
	flag.Parse()

	file, err := ioutil.ReadFile(configFileName)

	if err != nil {
		log.Fatalf("Could not read config file: %v", err)
	}

	if err := json.Unmarshal(file, &josuke.Config); err != nil {
		log.Fatalf("Could not parse json from config file")
	}

	go func(httpPort int) {
		http.HandleFunc("/", josuke.HttpHandle)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", httpPort), nil))
	}(*httpPort)

	s := &http.Server{
		Addr: fmt.Sprintf(":%d", *port),
		Handler: &josuke.Handler{
			Uri: *uri,
		},
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
