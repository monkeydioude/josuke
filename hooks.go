package josuke

import (
	"io"
	"io/ioutil"
	"fmt"
	"log"
	"net/http"
	"strings"
)


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

	githubSignature := req.Header.Get("x-hub-signature-256")
	if githubSignature == "" {
		log.Println("[ERR ] x-hub-signature-256 was empty in headers")
		return
	} else {
		log.Printf("[INFO] check signature: %s\n",  githubSignature)
	}

	
	githubEvent := req.Header.Get("x-github-event")
	if githubEvent == "" {
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
