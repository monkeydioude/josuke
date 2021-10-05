package josuke

import (
	"io"
	"io/ioutil"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// retrieve hook from josuke
func (j *Josuke) getHook(name string) *Hook {
	log.Printf("[INFO] hooks count: %s\n", string(len(*j.Hooks)))
	for _, hook := range *j.Hooks {
		log.Printf("[INFO] about to loop hook: %s\n", hook.Name)
		if hook.matches(name) {
			return hook
		}
	}
	return nil
}
/*
func getHook(hooks *[]*Hook, name string) *Hook {
	log.Printf("[INFO] about to loop hook\n")
	for _, hook := range *hooks {
		log.Printf("[INFO] loop hook: %s\n", hook.Name)

		if hook.matches(name) {
			return hook
		}
	}
	return nil
}
*/
// GithubRequest handles github's webhook triggers
func (j *Josuke) GithubRequest(rw http.ResponseWriter, req *http.Request) {
	log.Printf("[INFO] Caught call from GitHub %+v\n", req.URL)
	defer req.Body.Close()

	buf := new(strings.Builder)
	_, err := io.Copy(buf, req.Body)
	if err != nil {
		log.Fatal(err)
	}
	s := buf.String()

	bodyReader := ioutil.NopCloser(strings.NewReader(s))

	if j.Debug {
		log.Println("[DBG ] start body ====")
		fmt.Println(s)
		log.Println("[DBG ] end body ====")
	}

	payload, err := fetchPayload(bodyReader)

	if err != nil {
		log.Printf("[ERR ] Could not fetch Payload. Reason: %s", err)
		return
	}

	githubEvent := req.Header.Get("x-github-event")
	if githubEvent == "" {
		log.Println("[ERR ] x-github-event was empty in headers")
		return
	}

	githubSignature := req.Header.Get("x-hub-signature-256")
	if githubSignature == "" {
		log.Println("[ERR ] x-hub-signature-256 was empty in headers")
		return
	}

	log.Printf("[INFO] check signature: %s\n",  githubSignature)
	// FIXME : should aleady be present when calling this method.
	// hardcoded name
	hook := j.getHook("github")

	if hook == nil {
		log.Println("[ERR ] cannot find hook for secret")
		return
	}
	log.Printf("[INFO] hook secret: %s\n", hook.Secret)

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

// BitbucketRequest handles github's webhook triggers
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
