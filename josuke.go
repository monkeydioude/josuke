package josuke

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Defines a log level
type LogLevel int

// Available log levels
const (
	TraceLevel LogLevel = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
)

var name2logLevel = map[string]LogLevel{
	"TRACE": TraceLevel,
	"DEBUG": DebugLevel,
	"INFO":  InfoLevel,
	"WARN":  WarnLevel,
	"ERROR": ErrorLevel,
}

func parseLogLevel(value string) (LogLevel, bool) {
	c, ok := name2logLevel[strings.ToUpper(value)]
	return c, ok
}

type Josuke struct {
	LogLevelName string `json:"logLevel" default:"INFO"`
	LogLevel     LogLevel
	Hooks        *[]*Hook `json:"hook"`
	Deployment   *[]*Repo `json:"deployment"`
	Host         string   `json:"host" default:"localhost"`
	Port         int      `json:"port" default:"8082"`
	Cert         string   `json:"cert"`
	Key          string   `json:"key"`
	Store        string   `json:"store" default:""`
}

func New(configFilePath string) (*Josuke, error) {
	file, err := ioutil.ReadFile(configFilePath)

	if err != nil {
		return nil, fmt.Errorf("could not read config file: %v", err)
	}

	j := &Josuke{}

	if err := json.Unmarshal(file, j); err != nil {
		return nil, errors.New("could not parse json from config file")
	}

	logLevel, ok := parseLogLevel(j.LogLevelName)
	if !ok {
		return nil, fmt.Errorf("could not parse the log level: %s", j.LogLevelName)
	}
	j.LogLevel = logLevel

	return j, nil
}

func (j *Josuke) LogEnabled(ll LogLevel) bool {
	return j.LogLevel <= ll
}

var keyholders = map[string]func(*Info) string{
	"%base_dir%": func(i *Info) string {
		return i.BaseDir
	},
	"%proj_dir%": func(i *Info) string {
		return i.ProjDir
	},
	"%html_url%": func(i *Info) string {
		return i.HtmlUrl
	},
	"%payload_path%": func(i *Info) string {
		return i.PayloadPath
	},
}

type Hook struct {
	// Optional command, takes precedence over deployment commands if set.
	// Only %payload_path% placeholder is available.
	Command     []string `json:"command"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Path        string   `json:"path"`
	Secret      string   `json:"secret"`
	SecretBytes []byte
}

// Repository represents the payload repository information
type Repository struct {
	Name    string `json:"full_name"`
	HtmlUrl string `json:"html_url"`
}

// Repo is built from github's json payload, mirroring dir data from config, branches & repo name
type Repo struct {
	Name     string   `json:"repo"`
	Branches []Branch `json:"branches"`
	BaseDir  string   `json:"base_dir"`
	ProjDir  string   `json:"proj_dir"`
}

// Matches repo names from payload and config
func (r Repo) matches(trial string) bool {
	return r.Name == trial
}

// Info contains various data about directory to deploy to and git's repo url
type Info struct {
	BaseDir     string
	ProjDir     string
	HtmlUrl     string
	PayloadPath string
}

// Branch mirrors config's branch section, containing branch Name & Actions linked to it
type Branch struct {
	Name    string   `json:"branch"`
	Actions []Action `json:"actions"`
}

// Matches a branch name using payload & concatenation of static "refs/heads/" + config's branch name
func (b Branch) matches(trial string) bool {
	return fmt.Sprintf("%s%s", staticRefPrefix, b.Name) == trial
}

// Action contains set of commands from config matching the type of action sent from github (if action is "push", then we do "these" commands)
type Action struct {
	Action   string     `json:"action"`
	Commands [][]string `json:"commands"`
}

// Executes the retrieved set of commands from config
func (a *Action) execute(i *Info) error {
	switchToDefaultUser()
	for _, command := range a.Commands {
		if err := ExecuteCommand(command, i); err != nil {
			return err
		}
	}

	switchToDefaultUser()
	return nil
}

// Matches an action type using github's payload & config's action type
func (a Action) matches(trial string) bool {
	return a.Action == trial
}

// Config mirrors our json config file, used to boot this deployer
// var Config []Repo
var staticRefPrefix = "refs/heads/"

func fetchPayload(r io.Reader) (*Payload, error) {
	payload := &Payload{}
	err := json.NewDecoder(r).Decode(payload)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func chdir(args []string) error {
	if err := os.Chdir(args[0]); err != nil {
		return fmt.Errorf("%s on \"%s\" directory", err.Error(), args[0])
	}
	return nil
}

func replaceKeyholders(args []string, i *Info) []string {
	for k, arg := range args {
		if fun, ok := keyholders[arg]; ok {
			args[k] = fun(i)
		}
	}
	return args
}

// ExecuteCommand execute a command and its args coming in a form of a slice of string, using Info
func ExecuteCommand(c []string, i *Info) error {
	if len(c) == 0 {
		return fmt.Errorf("empty command slice")
	}
	name := c[0]
	args := make([]string, len(c)-1)
	copy(args, c[1:])
	args = replaceKeyholders(args, i)

	log.Printf("[INFO] executing %s %+v\n", name, args)

	if name == "cd" {
		return chdir(args)
	}

	if yes, user := isSwitchUserCall(name); yes {
		return SwitchUser(user)
	}

	if name == "git" && args[0] == "clone" {
		if _, err := os.Stat(i.ProjDir); !os.IsNotExist(err) {
			return nil
		}
	}
	cmd := exec.Command(name, args...)

	if err := NativeExecuteCommand(cmd); err != nil {
		return fmt.Errorf("could not execute command %s %v: %s", name, args, err)
	}
	return nil
}
