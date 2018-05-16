package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

var keyholders = map[string]interface{}{
	"%base_dir%": func(i *Info) string {
		return i.BaseDir
	},
	"%proj_dir%": func(i *Info) string {
		return i.ProjDir
	},
	"%html_url%": func(i *Info) string {
		return i.HtmlUrl
	},
}

type Repository struct {
	Name    string `json:"full_name"`
	HtmlUrl string `json:"html_url"`
}

// Payload fetching useful data from github's json payload
type Payload struct {
	Ref        string `json:"ref"`
	Action     string
	HtmlUrl    string
	Repository Repository `json:"repository"`
}

// retrieve branch name from config using Payload and matching config's Repo
func (p *Payload) getBranch(r *Repo) *Branch {
	for _, branch := range r.Branches {
		if branch.matches(p.Ref) {
			return &branch
		}
	}
	return nil
}

// retrieve repo from config using Paylaod
func (p *Payload) getRepo() *Repo {
	for _, repo := range Config {
		if repo.matches(p.Repository.Name) {
			return &repo
		}
	}
	return nil
}

// Process of retrieving deploy information from github payload
func (p *Payload) getDeployAction() (*Action, *Info) {
	repo := p.getRepo()
	if repo == nil {
		fmt.Println("Could not match any repo in config file. We'll just do nothing.")
		return nil, nil
	}
	branch := p.getBranch(repo)
	if branch == nil {
		fmt.Println("Could not find any matching branch. We'll just do nothing.")
		return nil, nil
	}
	// ref = fmt.Sprintf("%s%s", staticRefPrefix, )
	action := p.getAction(branch)
	if action == nil {
		fmt.Println("Could not find any matchin action. We'll just do nothing.")
		return nil, nil
	}
	return action, &Info{
		BaseDir: repo.BaseDir,
		ProjDir: repo.ProjDir,
		HtmlUrl: p.Repository.HtmlUrl,
	}
}

// retrieve action from config using Payload and matching config's branch
func (p *Payload) getAction(b *Branch) *Action {
	for _, action := range b.Actions {
		if action.matches(p.Action) {
			return &action
		}
	}
	return nil
}

// Repo is built from github's json payload, mirroring dir data from config, branches & repo name
type Repo struct {
	Name     string   `json:"repo"`
	Branches []Branch `json:"branches"`
	BaseDir  string   `json:"base_dir"`
	ProjDir  string   `json:"proj_dir"`
}

// Matches repo names from payload and config
func (r Repo) matches(trial string) bool {
	return r.Name == trial
}

// Info contains mixed data about repertory to deploy in and git's repo url
type Info struct {
	BaseDir string
	ProjDir string
	HtmlUrl string
}

// Branch mirrors config's branch section, containing branch Name & Actions linked to it
type Branch struct {
	Name    string   `json:"branch"`
	Actions []Action `json:"actions"`
}

// Matches a branch name using payload & concatenation of static "refs/heads/" + config's branch name
func (b Branch) matches(trial string) bool {
	return fmt.Sprintf("%s%s", staticRefPrefix, b.Name) == trial
}

// Action contains set of commands from config matching the type of action sent from github (if action is "push", then we do "these" commands)
type Action struct {
	Action   string     `json:"action"`
	Commands [][]string `json:"commands"`
}

// Executes the retrived set of commands from config
func (a *Action) execute(i *Info) error {
	for _, command := range a.Commands {
		if err := ExecuteCommand(command, i); err != nil {
			return err
		}
	}

	return nil
}

// Matches an action type using github's payload & config's action type
func (a Action) matches(trial string) bool {
	return a.Action == trial
}

// Config mirrors our json config file, used to boot this deployer
var Config []Repo
var staticRefPrefix = "refs/heads/"

func fetchPayload(r io.Reader) *Payload {
	payload := new(Payload)
	err := json.NewDecoder(r).Decode(payload)
	if err != nil {
		panic(err)
	}
	return payload
}

func chdir(args []string, i *Info) error {
	args = replaceKeyholders(args, i)
	if err := os.Chdir(args[0]); err != nil {
		return fmt.Errorf("%s on \"%s\" directory", err.Error(), args[0])
	}
	return nil
}

func replaceKeyholders(args []string, i *Info) []string {
	for k, arg := range args {
		if val, ok := keyholders[arg]; ok {
			args[k] = val.(func(*Info) string)(i)
		}
	}
	return args
}

// ExecuteCommand execute a command and its args coming in a form of a slice of string, using Info
func ExecuteCommand(c []string, i *Info) error {
	if len(c) == 0 {
		return fmt.Errorf("Empy command slice")
	}
	name := c[0]
	var args []string
	if len(c) > 1 {
		args = c[1:len(c)]
	}
	if name == "cd" {
		return chdir(args, i)
	}

	if name == "git" && args[0] == "clone" {
		if _, err := os.Stat(i.ProjDir); !os.IsNotExist(err) {
			return nil
		}
	}
	args = replaceKeyholders(args, i)
	cmd := exec.Command(name, args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Could not execute command %s %v: %s", name, args, err.Error())
	}
	return nil
}

// Request handle github's webhook triggers
func GithubRequest(rw http.ResponseWriter, req *http.Request) {
	var githubEvent string
	payload := fetchPayload(req.Body)

	defer req.Body.Close()

	if githubEvent = req.Header.Get("x-github-event"); githubEvent == "" {
		return
	}

	payload.Action = githubEvent

	action, info := payload.getDeployAction()
	if action == nil {
		return
	}

	if err := action.execute(info); err != nil {
		fmt.Println(err.Error())
	}
}

func BitbucketRequest(rw http.ResponseWriter, req *http.Request) {
	payload := bitbucketToPayload(req.Body)

	defer req.Body.Close()

	action, info := payload.getDeployAction()
	if action == nil {
		return
	}

	if err := action.execute(info); err != nil {
		fmt.Println(err.Error())
	}
}
