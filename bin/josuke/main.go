package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/monkeydioude/josuke"
)

// findOutProtocolHandler defines which handler, with respect to the protocol,
// should be used, from a josuke.Josuke struct
func findOutProtocolHandler(j *josuke.Josuke) (string, func() error) {
	p := fmt.Sprintf("%s:%d", j.Host, j.Port)
	if j.Key == "" {
		return "http", func() error {
			return http.ListenAndServe(p, nil)
		}
	}

	return "https", func() error {
		return http.ListenAndServeTLS(p, j.Cert, j.Key, nil)
	}
}

func main() {
	configFileName := flag.String("c", "config.json", "Path to config file")
	flag.Parse()

	j, err := josuke.New(*configFileName)

	if err != nil {
		log.Fatal("[ERR ] %s", err)
	}

	if *j.Hooks == nil || len(*j.Hooks) == 0 {
		log.Fatal("[ERR ] MUDA MUDA MUDA ! Josuke needs to handle at least one type of hook. See README.md for help")
	}

	for _, hook := range *j.Hooks {
		//log.Printf("[INFO] add hook %s: %s\n", hook.Name, hook.Path)
		if hook.Secret != "" && hook.SecretBytes == nil {
			hook.SecretBytes = []byte(hook.Secret)
		}

		hh := &josuke.HookHandler{}
		hh.Josuke = j
		hh.Hook = hook

		if hook.Type == "github" {
			hh.Handler = hh.GithubRequest
			log.Println("[INFO] Gureto daze 8), handling Github hooks")
		} else if hook.Type == "bitbucket" {
			hh.Handler = hh.BitbucketRequest
			log.Println("[INFO] Gureto daze 8), handling Bitbucket hooks")
		} else if hook.Type == "gogs" {
			hh.Handler = hh.GogsRequest
			log.Println("[INFO] Gureto daze 8), handling Gogs hooks")
		} else {
			log.Fatal(fmt.Sprintf("[ERR ] Oh, My, God ! Josuke does not know this type of hook: %s. See README.md for help", hook.Type))
		}
		http.HandleFunc(hook.Path, hh.Handler)
	}

	protocol, handler := findOutProtocolHandler(j)
	log.Printf("[INFO] Listening %s://%s:%d\n", protocol, j.Host, j.Port)
	log.Fatal(handler())
}
