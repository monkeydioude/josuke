package josuke

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

type Payload struct {
	Ref        string `json:"ref"`
	Action     string
	HtmlUrl    string `json:"html_url"`
	Repository struct {
		Name string `json:"full_name"`
	} `json:"repository"`
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
	for _, repo := range Config {
		if repo.matches(p.Repository.Name) {
			return &repo
		}
	}
	return nil
}

func (p *Payload) getDeployAction() (*Action, *Info) {
	repo := p.getRepo()
	if repo == nil {
		fmt.Println("Could not match any repo in config file. We'll just do nothing.")
		return nil, nil
	}
	branch := p.getBranch(repo)
	if repo == nil {
		fmt.Println("Could not find any matching branch. We'll just do nothing.")
		return nil, nil
	}
	// ref = fmt.Sprintf("%s%s", staticRefPrefix, )
	action := p.getAction(branch)
	if action == nil {
		fmt.Println("Could not find any matchin action. We'll just do nothing.")
		return nil, nil
	}
	repo.Info.HtmlUrl = p.HtmlUrl
	return action, &repo.Info
}

func (p *Payload) getAction(b *Branch) *Action {
	for _, action := range b.Actions {
		if action.matches(p.Action) {
			return &action
		}
	}
	return nil
}

type Repo struct {
	Name     string   `json:"repo"`
	Branches []Branch `json:"branches"`
	Info     Info
}

func (r Repo) matches(trial string) bool {
	return r.Name == fmt.Sprintf("%s%s", staticRefPrefix, trial)
}

type Info struct {
	BaseDir string `json:"base_dir"`
	ProjDir string `json:"proj_dir"`
	HtmlUrl string
}

type Branch struct {
	Name    string   `json:"branch"`
	Actions []Action `json:"actions"`
}

func (b Branch) matches(trial string) bool {
	return b.Name == trial
}

type Action struct {
	Action   string     `json:"action"`
	Commands [][]string `json:"commands"`
}

func (a *Action) execute(i *Info) error {
	os.Chdir(i.BaseDir)
	if _, err := os.Stat(i.ProjDir); os.IsNotExist(err) {
		ExecuteCommand([]string{"git", "clone", i.HtmlUrl})
	}
	os.Chdir(i.ProjDir)
	for _, command := range a.Commands {
		ExecuteCommand(command)
	}

	return nil
}

func (a Action) matches(trial string) bool {
	return a.Action == trial
}

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

func ExecuteCommand(c []string) error {
	if len(c) == 0 {
		return fmt.Errorf("Empy command slice")
	}
	name := c[0]
	var args []string
	if len(c) > 1 {
		args = c[1:len(c)]
	}
	cmd := exec.Command(name, args...)
	cmd.Run()
	return nil
}

func Request(rw http.ResponseWriter, req *http.Request) {
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

	if action.execute(info) != nil {
		fmt.Println("could not execute action")
	}
}
