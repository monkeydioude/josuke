package josuke

import (
	"io"
	"io/ioutil"
//    "encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// retrieve hook from josuke
func (j *Josuke) getHook(name string) *Hook {
	for _, hook := range *j.Hooks {
		//log.Printf("[INFO] about to loop hook: %s\n", hook.Name)
		if hook.matches(name) {
			return hook
		}
	}
	return nil
}

func getBody(reader io.Reader, debug bool) (string, error) {
	buf := new(strings.Builder)
	_, err := io.Copy(buf, reader)
	if err != nil {
		return "", err
	}
	s := buf.String()

	if debug {
		log.Printf("[DBG ] start body %d ====\n", len(s))
		fmt.Println(s)
		log.Println("[DBG ] end body ====")
		//log.Println(hex.EncodeToString([]byte(s)))
		//log.Println("[DBG ] end body as hex ====")
	}
	return s, nil
}

func (j *Josuke) GenericRequest(
	rw http.ResponseWriter,
	req *http.Request,
	scmType string,
	eventHeaderName string,
	signatureHeaderName string) {

	log.Printf("[INFO] Caught call from %s %+v\n", scmType, req.URL)
	defer req.Body.Close()

	scmEvent := req.Header.Get(eventHeaderName)
	if scmEvent == "" {
		log.Printf("[ERR ] %s was empty in headers\n", eventHeaderName)
		return
	}

	hook := j.getHook(scmType)

	if hook == nil {
		log.Println("[ERR ] cannot find hook for secret")
		return
	}

	s, err := getBody(req.Body, j.Debug)
	if err != nil {
		log.Printf("[ERR ] Could not read body. Reason: %s", err)
		return
	}

	if hook.SecretBytes != nil {
		requestSignature := req.Header.Get(signatureHeaderName)
		if requestSignature == "" {
			log.Printf("[ERR ] %s was empty in headers\n", signatureHeaderName)
			return
		}

		signature := hmacSha256(hook.SecretBytes, s)
		//log.Printf("[INFO] payload signature: %s\n", signature)
		// TODO ConstantTimeCompare to not leak information
		if requestSignature != signature {
			log.Printf("[ERR ] payload signature does not match:\n  request  %s\n  expected %s\n", requestSignature, signature)
			return
		}
	}

	bodyReader := ioutil.NopCloser(strings.NewReader(s))

	payload, err := fetchPayload(bodyReader)

	if err != nil {
		log.Printf("[ERR ] Could not fetch payload. Reason: %s", err)
		return
	}

	payload.Action = scmEvent

	action, info := payload.getDeployAction(j.Deployment)
	if action == nil {
		log.Println("[ERR ] Could not retrieve any action")
		return
	}

	if err := action.execute(info); err != nil {
		log.Printf("[ERR ] Could not execute action. Reason: %s", err)
	}
}

// GogsRequest handles gogs' webhook triggers
func (j *Josuke) GogsRequest(rw http.ResponseWriter, req *http.Request) {

	scmType := "gogs"
	// X-Gogs-Delivery
	eventHeaderName := "x-gogs-event"
	signatureHeaderName := "x-gogs-signature"

	j.GenericRequest(rw, req, scmType, eventHeaderName, signatureHeaderName)
}

// GithubRequest handles github's webhook triggers
func (j *Josuke) GithubRequest(rw http.ResponseWriter, req *http.Request) {

	scmType := "github"
	eventHeaderName := "x-github-event"
	// Could be x-hub-signature for older servers.
	signatureHeaderName := "x-hub-signature-256"

	j.GenericRequest(rw, req, scmType, eventHeaderName, signatureHeaderName)
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
