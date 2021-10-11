package josuke

import (
	"io"
	"io/ioutil"
//    "encoding/hex"
	"fmt"
	"log"
	"net/http"
	"math/rand"
	"strings"
	"time"
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

func randomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(65 + rand.Intn(25))  //A=65 and Z = 65+25
	}
	return string(bytes)
}

// Contains the hook and the HTTP response handler.
type HookHandler struct {
	Josuke *Josuke
	Hook *Hook
	Handler func(http.ResponseWriter, *http.Request)
}

// Checks request HMAC 256 from a HTTP header and runs the action.
func (hh *HookHandler) GenericRequest(
	rw http.ResponseWriter,
	req *http.Request,
	eventHeaderName string,
	signatureHeaderName string) {

	log.Printf("[INFO] Caught call from %s %+v\n", hh.Hook.Type, req.URL)
	defer req.Body.Close()

	scmEvent := req.Header.Get(eventHeaderName)
	if scmEvent == "" {
		log.Printf("[ERR ] %s was empty in headers\n", eventHeaderName)
		return
	}

	payloadContent, err := getBody(req.Body, hh.Josuke.Debug)
	if err != nil {
		log.Printf("[ERR ] Could not read body. Reason: %s", err)
		return
	}

	if hh.Hook.SecretBytes != nil {
		requestSignature := req.Header.Get(signatureHeaderName)
		if requestSignature == "" {
			log.Printf("[ERR ] %s was empty in headers\n", signatureHeaderName)
			return
		}

		signature := hmacSha256(hh.Hook.SecretBytes, payloadContent)
		//log.Printf("[INFO] payload signature: %s\n", signature)
		// TODO use ConstantTimeCompare to avoid leaking information.
		if requestSignature != signature {
			log.Printf("[ERR ] payload signature does not match:\n  request  %s\n  expected %s\n", requestSignature, signature)
			return
		}
	}

	bodyReader := ioutil.NopCloser(strings.NewReader(payloadContent))

	var payloadPath string
	if hh.Josuke.Store != "" {

		t := time.Now().UTC()
		//buf := []rune(t.Format(time.RFC3339))
		dt := strings.ReplaceAll(t.Format(time.RFC3339), ":", "")
		//strings.Replace(
		payloadPath = hh.Josuke.Store + "/" + hh.Hook.Name + "." + dt + "." + randomString(6) + ".json"
		err = ioutil.WriteFile(payloadPath, []byte(payloadContent), 0664)
		if err != nil {
			log.Printf("[ERR ] cannot create the payload file: %s", err)
			return
		}
		log.Printf("[INFO] store payload to %s\n", payloadPath)
	} else {
		payloadPath = ""
	}

	payload, err := fetchPayload(bodyReader)

	if err != nil {
		log.Printf("[ERR ] Could not fetch payload. Reason: %s", err)
		return
	}

	payload.Action = scmEvent

	action, info := payload.getDeployAction(hh.Josuke.Deployment, payloadPath)

	if action == nil {
		log.Println("[ERR ] Could not retrieve any action")
		return
	}

	if err := action.execute(info); err != nil {
		log.Printf("[ERR ] Could not execute action. Reason: %s", err)
	}
}

// GogsRequest handles gogs' webhook triggers
func (hh *HookHandler) GogsRequest(rw http.ResponseWriter, req *http.Request) {

	// X-Gogs-Delivery
	eventHeaderName := "x-gogs-event"
	signatureHeaderName := "x-gogs-signature"

	hh.GenericRequest(rw, req, eventHeaderName, signatureHeaderName)
}

// GithubRequest handles github's webhook triggers
func (hh *HookHandler) GithubRequest(rw http.ResponseWriter, req *http.Request) {

	eventHeaderName := "x-github-event"
	// Could be x-hub-signature for older servers.
	signatureHeaderName := "x-hub-signature-256"

	hh.GenericRequest(rw, req, eventHeaderName, signatureHeaderName)
}

// BitbucketRequest handles github's webhook triggers
func (hh *HookHandler) BitbucketRequest(rw http.ResponseWriter, req *http.Request) {
	log.Printf("[INFO] Caught call from BitBucket %+v\n", req.URL)
	payload := bitbucketToPayload(req.Body)

	defer req.Body.Close()

	// TODO : implement payload path for BitbucketRequest
	payloadPath := ""
	action, info := payload.getDeployAction(hh.Josuke.Deployment, payloadPath)
	if action == nil {
		log.Println("[ERR ] Could not retrieve any action")
		return
	}

	if err := action.execute(info); err != nil {
		log.Printf("[ERR ] Could not execute action. Reason: %s", err)
	}
}
