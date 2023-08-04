package josuke

import (
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// HookDef defines a type of hook, usually a Source Code Management (Gogs, GitHub, BitBucket).
type HookDef struct {
	Name    string
	Title   string
	Handler func(http.ResponseWriter, *http.Request)
}

func getBody(reader io.Reader, logLevel LogLevel) (string, error) {
	buf := new(strings.Builder)
	_, err := io.Copy(buf, reader)
	if err != nil {
		return "", err
	}
	s := buf.String()

	if logLevel <= DebugLevel {
		log.Printf("[DBG ] start body %d ====\n", len(s))
		fmt.Println(s)
		log.Println("[DBG ] end body ====")
		if logLevel <= TraceLevel {
			log.Println("[TRAC] start body as hex ====")
			log.Println(hex.EncodeToString([]byte(s)))
			log.Println("[TRAC] end body as hex ====")
		}
	}
	return s, nil
}

func randomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(65 + rand.Intn(25)) //A=65 and Z = 65+25
	}
	return string(bytes)
}

// storePayload write the payload content if the store directory path is set on the hook handler.
// Returns the payload path or an empty string if the payload is not stored.
func storePayload(payloadContent string, hh *HookHandler) (string, error) {
	if hh.Josuke.Store == "" {
		return "", nil
	}

	t := time.Now().UTC()
	dt := strings.ReplaceAll(t.Format(time.RFC3339), ":", "")
	payloadPath := hh.Josuke.Store + "/" + hh.Hook.Name + "." + dt + "." + randomString(6) + ".json"
	err := ioutil.WriteFile(payloadPath, []byte(payloadContent), 0664)
	if err != nil {
		return "", fmt.Errorf("cannot write the payload: %s", err)
	}
	log.Printf("[INFO] store payload to %s\n", payloadPath)
	return payloadPath, nil
}

// HookHandler contains the hook definition, the reified hook and the HTTP response handler.
type HookHandler struct {
	Josuke  *Josuke
	Hook    *Hook
	HookDef *HookDef
}

// GenericRequest checks request HMAC 256 from a HTTP header and runs the action.
func (hh *HookHandler) GenericRequest(
	rw http.ResponseWriter,
	req *http.Request,
	eventHeaderName string,
	signatureHeaderName string,
) {

	log.Printf("[INFO] Caught call from %s %+v\n", hh.Hook.Type, req.URL)
	defer req.Body.Close()

	hookEvent := req.Header.Get(eventHeaderName)
	if hookEvent == "" {
		log.Printf("[ERR ] %s was empty in headers\n", eventHeaderName)
		return
	}

	payloadContent, err := getBody(req.Body, hh.Josuke.LogLevel)
	if err != nil {
		log.Printf("[ERR ] Could not read body. Reason: %s", err)
		return
	}

	if hh.Hook.SecretBytes != nil {
		requestSignature := req.Header.Get(signatureHeaderName)
		digestName := "sha256"
		if requestSignature == "" {
			log.Printf("[ERR ] %s was empty in headers\n", signatureHeaderName)
			return
		}
		equalIndex := strings.Index(requestSignature, "=")
		if equalIndex > -1 {
			digestName = requestSignature[:equalIndex]
			requestSignature = requestSignature[equalIndex+1:]
		}

		// TODO one hash sha256 as of now. Could have a dictionary: digest name to digest method.
		if digestName != "sha256" {
			log.Printf("[ERR ] payload signature digest not handled: %s\n", digestName)
			return
		}

		signature := hmacSha256(hh.Hook.SecretBytes, payloadContent)
		if hh.Josuke.LogEnabled(TraceLevel) {
			log.Printf("[TRAC] payload signature: %s\n", signature)
		}
		// TODO use ConstantTimeCompare to avoid leaking information.
		if requestSignature != signature {
			log.Printf("[ERR ] payload signature does not match:\n  request  %s\n  expected %s\n", requestSignature, signature)
			return
		}
	}

	payloadPath, err := storePayload(payloadContent, hh)
	if err != nil {
		log.Printf("[ERR ] cannot store the payload: %s", err)
	}

	bodyReader := ioutil.NopCloser(strings.NewReader(payloadContent))
	payload, err := fetchPayload(bodyReader)

	if err != nil {
		log.Printf("[ERR ] could not fetch payload: %s", err)
		return
	}

	payload.Action = hookEvent

	hookActions := hh.getHookActions(payload, payloadPath)
	if len(hookActions) == 0 {
		log.Println("[ERR ] Could not retrieve any action")
		return
	}

	for _, ha := range hookActions {
		if ha.Action == nil {
			continue
		}
		if err := ha.Action.execute(ha.Info); err != nil {
			log.Printf("[ERR ] Could not execute action. Reason: %s", err)
		}
	}
}

type HookAction struct {
	Action *Action
	Info   *Info
}

// Returns either the hook command if present, or a deployment command.
func (hh *HookHandler) getHookActions(payload *Payload, payloadPath string) []HookAction {
	var hookActions []HookAction

	// hook action gets triggered first before deploy
	if hh.Hook.Command != nil && len(hh.Hook.Command) > 0 {
		if hh.Josuke.LogEnabled(TraceLevel) {
			log.Println("[TRAC] hook action")
		}
		hookActions = append(
			hookActions,
			HookAction{
				Action: &Action{
					Action:   "hook",
					Commands: [][]string{hh.Hook.Command},
				},
				Info: &Info{
					BaseDir:      "",
					ProjDir:      "",
					HtmlUrl:      "",
					PayloadHook:  hh.Hook.Name,
					PayloadPath:  payloadPath,
					PayloadEvent: payload.Action,
				},
			})
	}
	if hh.Josuke.LogEnabled(TraceLevel) {
		log.Println("[TRAC] hook action from deployment")
	}
	if hh.Josuke.Deployment == nil {
		return hookActions
	}
	action, info := payload.getDeployAction(hh.Josuke.Deployment, payloadPath, hh.Hook.Name, payload.Action)

	// No deployment found
	if action == nil {
		return hookActions
	}

	// deployment action gets triggered second after hook action
	return append(
		hookActions,
		HookAction{
			Action: action,
			Info:   info,
		})
}

// WebhookRequest handles generic webhook triggers
func (hh *HookHandler) WebhookRequest(rw http.ResponseWriter, req *http.Request) {

	eventHeaderName := "x-webhook-event"
	signatureHeaderName := "x-webhook-signature"

	hh.GenericRequest(rw, req, eventHeaderName, signatureHeaderName)
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

	eventHeaderName := "x-event-key"

	hookEvent := req.Header.Get(eventHeaderName)
	if hookEvent == "" {
		log.Printf("[ERR ] %s was empty in headers\n", eventHeaderName)
		return
	}

	defer req.Body.Close()

	payload, err := bitbucketToPayload(req.Body, hookEvent)
	if err != nil {
		log.Printf("[ERR ] Could not read body. Reason: %s", err)
		return
	}

	// TODO : implement payload path for BitbucketRequest
	payloadPath := ""
	action, info := payload.getDeployAction(hh.Josuke.Deployment, payloadPath, hh.Hook.Name, payload.Action)

	if action == nil {
		log.Println("[ERR ] Could not retrieve any action")
		return
	}

	if err := action.execute(info); err != nil {
		log.Printf("[ERR ] Could not execute action. Reason: %s", err)
	}
}

var type2hookDef = map[string]func(hh *HookHandler) *HookDef{
	"bitbucket": func(hh *HookHandler) *HookDef {
		return &HookDef{
			Name:    "bitbucket",
			Title:   "Bitbucket",
			Handler: hh.BitbucketRequest,
		}
	},
	"github": func(hh *HookHandler) *HookDef {
		return &HookDef{
			Name:    "github",
			Title:   "GitHub",
			Handler: hh.GithubRequest,
		}
	},
	"gogs": func(hh *HookHandler) *HookDef {
		return &HookDef{
			Name:    "gogs",
			Title:   "Gogs",
			Handler: hh.GogsRequest,
		}
	},
	"webhook": func(hh *HookHandler) *HookDef {
		return &HookDef{
			Name:    "webhook",
			Title:   "Webhook",
			Handler: hh.WebhookRequest,
		}
	},
}

func parseHookDef(hh *HookHandler, t string) *HookDef {
	if fun, ok := type2hookDef[t]; ok {
		return fun(hh)
	}
	return nil
}

// NewHookHandler constructs a hook handler with josuke and a hook definition.
func NewHookHandler(j *Josuke, h *Hook) (*HookHandler, error) {
	hh := &HookHandler{
		Josuke: j,
		Hook:   h,
	}
	hookDef := parseHookDef(hh, h.Type)
	if nil == hookDef {
		return nil, fmt.Errorf("oh, my, god ! Josuke does not know this type of hook: %s. See README.md for help", h.Type)
	}
	hh.HookDef = hookDef
	return hh, nil
}
