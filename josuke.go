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
	Action     string
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
	Action   string   `json:"action"`
	Commands Commands `json:"commands"`
	Dir      string
}

type Commands [][]string

var config Config
var staticRefPrefix = "refs/heads/"

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

func (a Action) matches(trial string) bool {
	return a.Action == trial
}

func (b Branch) matches(trial string) bool {
	return b.Name == trial
}

func (r Repo) matches(trial string) bool {
	return r.Name == fmt.Sprintf("%s%s", staticRefPrefix, trial)
}

func (p *Payload) getAction(b *Branch) *Action {
	for _, action := range b.Actions {
		if action.matches(p.Action) {
			return &action
		}
	}
	return nil
}

func (p *Payload) getBranch(r *Repo) *Branch {
	for _, branch := range r.Branches {
		if branch.matches(p.Ref) {
			return &branch
		}
	}
	return nil
}

func (p *Payload) getRepo() *Repo {
	for _, repo := range config {
		if repo.matches(p.Repository.Name) {
			return &repo
		}
	}
	return nil
}

func (p *Payload) getDeployData() *Action {
	repo := p.getRepo()
	if repo == nil {
		fmt.Println("Could not match any repo in config file. We'll just do nothing.")
		return nil
	}
	branch := p.getBranch(repo)
	if repo == nil {
		fmt.Println("Could not find any matching branch. We'll just do nothing.")
		return nil
	}
	// ref = fmt.Sprintf("%s%s", staticRefPrefix, )
	return p.getAction(branch)
}

func (c *Action) deploy() {
}

func request(rw http.ResponseWriter, req *http.Request) {
	payload := new(Payload)
	err := json.NewDecoder(req.Body).Decode(payload)
	if err != nil {
		panic(err)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(req.Body)
	defer req.Body.Close()
	action := req.Header.Get("x-github-event")
	if action == "" {
		return
	}
	payload.Action = action
	data := payload.getDeployData()
	if data == nil {
		return
	}
	data.deploy()
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
	http.HandleFunc("/deploy", request)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
