package josuke

import (
	"log"
)

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

// retrieve repo from config using Payload
func (p *Payload) getRepo(deployment *[]*Repo) *Repo {
	for _, repo := range *deployment {
		if repo.matches(p.Repository.Name) {
			return repo
		}
	}
	return nil
}

// Process of retrieving deploy information from github payload
func (p *Payload) getDeployAction(deployment *[]*Repo, payloadPath string) (*Action, *Info) {
	repo := p.getRepo(deployment)
	if repo == nil {
		log.Println("[WARN] Could not match any repo in config file. We'll just do nothing.")
		return nil, nil
	}
	branch := p.getBranch(repo)
	if branch == nil {
		log.Println("[WARN] Could not find any matching branch. We'll just do nothing.")
		return nil, nil
	}
	// ref = fmt.Sprintf("%s%s", staticRefPrefix, )
	action := p.getAction(branch)
	if action == nil {
		log.Println("[WARN] Could not find any matching action. We'll just do nothing.")
		return nil, nil
	}
	return action, &Info{
		BaseDir: repo.BaseDir,
		ProjDir: repo.ProjDir,
		HtmlUrl: p.Repository.HtmlUrl,
		PayloadPath: payloadPath,
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
