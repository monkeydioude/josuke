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
		log.Fatal("[ERR ] ", err)
	}

	if *j.Hooks == nil || len(*j.Hooks) == 0 {
		log.Fatal("[ERR ] MUDA MUDA MUDA ! Josuke needs to handle at least one type of hook. See README.md for help")
	}

	for _, hook := range *j.Hooks {
		if j.LogEnabled(josuke.TraceLevel) {
			log.Printf("[TRAC] add hook %s (%s): %s\n", hook.Name, hook.Type, hook.Path)
		}
		if hook.Secret != "" && hook.SecretBytes == nil {
			hook.SecretBytes = []byte(hook.Secret)
		}

		hh, err := josuke.NewHookHandler(j, hook)
		if err != nil {
			log.Fatal("[ERR ] ", err)
		}

		if j.LogEnabled(josuke.InfoLevel) {
			log.Printf("[INFO] Gureto daze 8), handling %s hook %s\n", hh.Scm.Title, hh.Hook.Name)
		}

		if j.LogEnabled(josuke.DebugLevel) && nil != hh.Hook.Command && 0 > len(hh.Hook.Command) {
			log.Println("[DBG ] hook command: ", hh.Hook.Command)
		}
		http.HandleFunc(hook.Path, hh.Scm.Handler)
	}

	protocol, handler := findOutProtocolHandler(j)
	if j.LogEnabled(josuke.InfoLevel) {
		log.Printf("[INFO] Listening %s://%s:%d\n", protocol, j.Host, j.Port)
	}
	log.Fatal(handler())
}
