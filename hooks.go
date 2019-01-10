package josuke

import (
	"log"
	"net/http"
)

// Request handle github's webhook triggers
func (j *Josuke) GithubRequest(rw http.ResponseWriter, req *http.Request) {
	log.Printf("[INFO] Caught call from GitHub %+v\n", req.URL)
	var githubEvent string
	defer req.Body.Close()

	payload, err := fetchPayload(req.Body)

	if err != nil {
		log.Printf("[ERR ] Could not fetch Payload. Reason: %s", err)
		return
	}

	if githubEvent = req.Header.Get("x-github-event"); githubEvent == "" {
		log.Println("[ERR ] x-github-event was empty in headers")
		return
	}

	payload.Action = githubEvent

	action, info := payload.getDeployAction(j.Deployment)
	if action == nil {
		log.Println("[ERR ] Could not retrieve any action")
		return
	}

	if err := action.execute(info); err != nil {
		log.Printf("[ERR ] Could not execute action. Reason: %s", err)
	}
}

func (j *Josuke) BitbucketRequest(rw http.ResponseWriter, req *http.Request) {
	log.Printf("[INFO] Caught call from BitBucket %+v\n", req.URL)
	payload := bitbucketToPayload(req.Body)

	defer req.Body.Close()

	action, info := payload.getDeployAction(j.Deployment)
	if action == nil {
		log.Println("[ERR ] Could not retrieve any action")
		return
	}

	if err := action.execute(info); err != nil {
		log.Printf("[ERR ] Could not execute action. Reason: %s", err)
	}
}
