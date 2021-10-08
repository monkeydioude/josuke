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

// GithubRequest handles github's webhook triggers
func (j *Josuke) GithubRequest(rw http.ResponseWriter, req *http.Request) {
	log.Printf("[INFO] Caught call from GitHub %+v\n", req.URL)
	defer req.Body.Close()

	githubEvent := req.Header.Get("x-github-event")
	if githubEvent == "" {
		log.Println("[ERR ] x-github-event was empty in headers")
		return
	}

	requestSignature := req.Header.Get("x-hub-signature-256")
	if requestSignature == "" {
		log.Println("[ERR ] x-hub-signature-256 was empty in headers")
		return
	}

	buf := new(strings.Builder)
	_, err := io.Copy(buf, req.Body)
	if err != nil {
		log.Printf("[ERR ] Could not read Payload. Reason: %s", err)
		return
	}
	s := buf.String()

	if j.Debug {
		log.Printf("[DBG ] start body %d ====\n", len(s))
		fmt.Println(s)
		log.Println("[DBG ] end body ====")
		//log.Println(hex.EncodeToString([]byte(s)))
		//log.Println("[DBG ] end body as hex ====")
	}

	bodyReader := ioutil.NopCloser(strings.NewReader(s))

	//log.Printf("[INFO] check signature: %s\n",  requestSignature)
	// FIXME : should aleady be present when calling this method.
	// hardcoded name
	hook := j.getHook("github")

	if hook == nil {
		log.Println("[ERR ] cannot find hook for secret")
		return
	}

	if hook.SecretBytes == nil {
//		log.Printf("[INFO] hook secret: %s\n", hook.Secret)
//		secretBytes, err := hex.DecodeString(hook.Secret)
//		if err != nil {
//			log.Printf("[ERR ] cannot decode hex secret: %s\n", err)
//			return
//		}
//		hook.SecretBytes = secretBytes
		hook.SecretBytes = []byte(hook.Secret)
	}

	signature := hmacSha256(hook.SecretBytes, s)
	//log.Printf("[INFO] payload signature: %s\n", signature)
	// TODO ConstantTimeCompare to not leak information
	if requestSignature != signature {
		log.Printf("[ERR ] payload signature does not match:\n  request  %s\n  expected %s\n", requestSignature, signature)
		return
	}
		
	payload, err := fetchPayload(bodyReader)

	if err != nil {
		log.Printf("[ERR ] Could not fetch Payload. Reason: %s", err)
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
