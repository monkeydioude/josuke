package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
)

// Payload translation of github json into struct
type Payload struct {
	Ref string `json:"ref"`
}

type Repos []Repo

type Repo struct {
	Name     string   `json:"repo"`
	Branches []Branch `json:"branches`
}

type Branch struct {
	Name    string   `json:"branch"`
	Actions []Action `json:"actions"`
}

type Action struct {
	Action   string                `json:"action"`
	Commands []map[string][]string `json:"commands"`
}

func pushRequest(payload *Payload, rw http.ResponseWriter) error {
	reg, err := regexp.Compile("^.+/.+/(.+)$")

	if err != nil {
		return errors.New("Could not compile Regexp")
	}

	branches := reg.FindStringSubmatch(payload.Ref)

	if len(branches) <= 1 {
		return errors.New("Could not find ref branch")
	}
	os.Chdir("/var/www")
	//      branch := branches[1]
	cmdName := "git"
	cmdArgs := []string{"clone", "git@github.com:monkeydioude/donut.git"}
	cmd := exec.Command(cmdName, cmdArgs...)
	if _, err = os.Stat("donut"); os.IsNotExist(err) {
		cmd.Run()
	}
	// if err = cmd.Run(); err != nil {
	//      return errors.New("Could not Execute Command git")
	// }
	os.Chdir("donut")
	cmdName = "git"
	cmdArgs = []string{"fetch", "--all"}
	cmd = exec.Command(cmdName, cmdArgs...)
	cmd.Run()
	cmdName = "git"
	cmdArgs = []string{"checkout", "master"}
	cmd = exec.Command(cmdName, cmdArgs...)
	cmd.Run()
	cmdName = "git"
	cmdArgs = []string{"reset", "--hard", "origin/master"}
	cmd = exec.Command(cmdName, cmdArgs...)
	cmd.Run()
	cmdName = "make"
	cmdArgs = []string{}
	cmd = exec.Command(cmdName, cmdArgs...)
	cmd.Run()
	return nil
}

func deploy(rw http.ResponseWriter, req *http.Request) {
	payload := new(Payload)
	err := json.NewDecoder(req.Body).Decode(payload)
	if err != nil {
		panic(err)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(req.Body)
	defer req.Body.Close()
	if req.Header.Get("x-github-event") == "push" {
		file, e := ioutil.ReadFile("./config.json")
		if e != nil {
			panic(e)
		}
		fmt.Println(string(file))
		var repos Repos
		json.Unmarshal(file, &repos)
		fmt.Printf("%v\n", repos[0])
		// if pushRequest(payload, rw) != nil {
		// 	fmt.Print(err.Error())
		// }
	}
}

func main() {
	http.HandleFunc("/deploy", deploy)
	log.Fatal(http.ListenAndServe(":8082", nil))
}
