package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
)

type Payload struct {
	Ref        string `json:"ref"`
	Repository struct {
		Name string `json:"full_name"`
	} `json:"repository"`
}

type Config []Repo

type Repo struct {
	Name     string   `json:"repo"`
	Branches []Branch `json:"branches"`
	Dir      string   `json:"dir"`
}

type Branch struct {
	Name    string   `json:"branch"`
	Actions []Action `json:"actions"`
}

type Action struct {
	Action   string     `json:"action"`
	Commands [][]string `json:"commands"`
}

var config Config

func pushRequest(payload *Payload, rw http.ResponseWriter) error {
	reg, err := regexp.Compile("^.+/.+/(.+)$")

	if err != nil {
		return errors.New("Could not compile Regexp")
	}

	branches := reg.FindStringSubmatch(payload.Ref)

	if len(branches) <= 1 {
		return errors.New("Could not find ref branch")
	}
	os.Chdir("/var/www")
	//      branch := branches[1]
	cmdName := "git"
	cmdArgs := []string{"clone", "git@github.com:monkeydioude/donut.git"}
	cmd := exec.Command(cmdName, cmdArgs...)
	if _, err = os.Stat("donut"); os.IsNotExist(err) {
		cmd.Run()
	}
	// if err = cmd.Run(); err != nil {
	//      return errors.New("Could not Execute Command git")
	// }
	os.Chdir("donut")
	cmdName = "git"
	cmdArgs = []string{"fetch", "--all"}
	cmd = exec.Command(cmdName, cmdArgs...)
	cmd.Run()
	cmdName = "git"
	cmdArgs = []string{"checkout", "master"}
	cmd = exec.Command(cmdName, cmdArgs...)
	cmd.Run()
	cmdName = "git"
	cmdArgs = []string{"reset", "--hard", "origin/master"}
	cmd = exec.Command(cmdName, cmdArgs...)
	cmd.Run()
	cmdName = "make"
	cmdArgs = []string{}
	cmd = exec.Command(cmdName, cmdArgs...)
	cmd.Run()
	return nil
}

func deploy(rw http.ResponseWriter, req *http.Request) {
	payload := new(Payload)
	err := json.NewDecoder(req.Body).Decode(payload)
	if err != nil {
		panic(err)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(req.Body)
	defer req.Body.Close()
	fmt.Println(payload)
	if req.Header.Get("x-github-event") == "push" {
		// if pushRequest(payload, rw) != nil {
		// 	fmt.Print(err.Error())
		// }
	}
}

func main() {
	configFileName := flag.String("c", "config.json", "Path to config file")
	port := flag.Int("p", 8082, "Port server will listen to")
	file, err := ioutil.ReadFile(*configFileName)
	if err != nil {
		log.Fatalf("Could not read config file: %v", err)
	}
	if err := json.Unmarshal(file, &config); err != nil {
		log.Fatalf("Could not parse json from config file")
	}
	http.HandleFunc("/deploy", deploy)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
