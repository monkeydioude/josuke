package josuke

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"
)

// LogLevel defines the level of a log.
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

// Josuke is the main object, that contains the HTTP server configuration
// and the hook definitions.
type Josuke struct {
	LogLevel         LogLevel
	LogLevelName     string  `json:"logLevel" yaml:"logLevel"`
	Hooks            []*Hook `json:"hook" yaml:"hook"`
	Host             string  `json:"host" yaml:"host"`
	Port             int     `json:"port" yaml:"port"`
	Cert             string  `json:"cert" yaml:"cert"`
	Key              string  `json:"key" yaml:"key"`
	Store            string  `json:"store" yaml:"store"`
	HealthcheckRoute string  `json:"healthcheck_route,omitempty" yaml:"healthcheck_route,omitempty"`
	Deployment       []*Repo  `json:"deployment" yaml:"deployment"`
	QueueFile        string   `json:"queue_file,omitempty" yaml:"queue_file,omitempty"`
	JobsAPI          *JobsAPI `json:"jobs_api,omitempty" yaml:"jobs_api,omitempty"`

	queue *jobQueue
}

// New creates a josuke HTTP server that handles SCM webhooks.
func New(configFilePath string) (*Josuke, error) {
	j, err := parseConfig(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("could not parse config file: %v", err)
	}
	logLevel, ok := parseLogLevel(j.LogLevelName)
	if !ok {
		return nil, fmt.Errorf("could not parse the log level: %s", j.LogLevelName)
	}
	j.LogLevel = logLevel

	j.queue, err = newJobQueue(queueSize, j.QueueFile, runJob)
	if err != nil {
		return nil, fmt.Errorf("could not load the queue: %v", err)
	}
	return j, nil
}

// LogEnabled tests if the log statement should be printed for the given level.
func (j *Josuke) LogEnabled(ll LogLevel) bool {
	return j.LogLevel <= ll
}

// HandleHooks declare http Handlers from hooks defined
// in configuration's json given to binary
func (j *Josuke) HandleHooks() {
	if j.Hooks == nil {
		return
	}
	for _, hook := range j.Hooks {
		if j.LogEnabled(TraceLevel) {
			log.Printf("[TRAC] add hook %s (%s): %s\n", hook.Name, hook.Type, hook.Path)
		}
		if hook.Secret != "" && hook.SecretBytes == nil {
			hook.SecretBytes = []byte(hook.Secret)
		}

		hh, err := NewHookHandler(j, hook)
		if err != nil {
			log.Fatal("[ERR ] ", err)
		}

		if j.LogEnabled(InfoLevel) {
			log.Printf("[INFO] Gureto daze 8), handling %s hook %s\n", hh.HookDef.Title, hh.Hook.Name)
		}

		if j.LogEnabled(DebugLevel) && nil != hh.Hook.Command && len(hh.Hook.Command) > 0 {
			log.Println("[DBG ] hook command: ", hh.Hook.Command)
		}
		http.HandleFunc(hook.Path, hh.HookDef.Handler)
	}
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
	"%payload_event%": func(i *Info) string {
		return i.PayloadEvent
	},
	"%payload_hook%": func(i *Info) string {
		return i.PayloadHook
	},
}

// A Hook maps HTTP requests to local commands.
type Hook struct {
	// Optional command, takes precedence over deployment commands if set.
	// Only %payload_path%, %payload_event% and %payload_hook%
	// placeholders are available.
	Command     []string `json:"command,omitempty" yaml:"command,omitempty"`
	Name        string   `json:"name" yaml:"name"`
	Type        string   `json:"type" yaml:"type"`
	Path        string   `json:"path" yaml:"path"`
	Secret      string   `json:"secret" yaml:"secret"`
	SecretBytes []byte
}

// Repository represents the payload repository information
type Repository struct {
	Name    string `json:"full_name" yaml:"full_name"`
	HtmlUrl string `json:"html_url" yaml:"html_url"`
}

// Repo is built from github's json payload, mirroring dir data from config, branches & repo name
type Repo struct {
	Name     string   `json:"repo" yaml:"repo"`
	Branches []Branch `json:"branches" yaml:"branches"`
	BaseDir  string   `json:"base_dir" yaml:"base_dir"`
	ProjDir  string   `json:"proj_dir" yaml:"proj_dir"`
}

// Matches repo names from payload and config
func (r Repo) matches(trial string) bool {
	return r.Name == trial
}

// Info contains various data about directory to deploy to and git's repo url
type Info struct {
	BaseDir      string
	ProjDir      string
	HtmlUrl      string
	PayloadHook  string
	PayloadPath  string
	PayloadEvent string
}

// Branch mirrors config's branch section, containing branch Name & Actions linked to it
type Branch struct {
	Name    string   `json:"branch" yaml:"branch"`
	Actions []Action `json:"actions" yaml:"actions"`
}

// Matches a branch name using payload & concatenation of static "refs/heads/" + config's branch name
func (b Branch) matches(trial string) bool {
	return fmt.Sprintf("%s%s", staticRefPrefix, b.Name) == trial
}

// Action contains set of commands from config matching the type of action sent from github (if action is "push", then we do "these" commands)
type Action struct {
	Action   string     `json:"action" yaml:"action"`
	Commands [][]string `json:"commands" yaml:"commands"`
}

// Executes the retrieved set of commands from config
func (a *Action) execute(ctx context.Context, i *Info, logger *log.Logger) error {
	switchToDefaultUser()
	for _, command := range a.Commands {
		if err := ExecuteCommand(ctx, command, i, logger); err != nil {
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

// waitDelay is how long a stopped command has to exit before being killed.
const waitDelay = 10 * time.Second

// ExecuteCommand execute a command and its args coming in a form of a slice of string, using Info.
// The command output is written to logger line by line, while the command runs.
// The command is stopped if ctx is cancelled.
func ExecuteCommand(ctx context.Context, c []string, i *Info, logger *log.Logger) error {
	if len(c) == 0 {
		return fmt.Errorf("empty command slice")
	}
	name := c[0]
	// Clone so placeholders are not replaced in the config itself, they would be stuck to the first request values.
	args := replaceKeyholders(slices.Clone(c[1:]), i)

	logger.Printf("[INFO] executing %s %+v\n", name, args)

	if name == "cd" {
		return chdir(args)
	}

	if yes, user := isSwitchUserCall(name); yes {
		return SwitchUser(user)
	}

	if name == "git" && len(args) > 0 && args[0] == "clone" {
		if _, err := os.Stat(i.ProjDir); !os.IsNotExist(err) {
			logger.Printf("[INFO] %s already exists, skipping clone\n", i.ProjDir)
			return nil
		}
	}
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = os.Environ()
	// Same writer for both streams so they are merged in order:
	// tools like git and docker report their progress on stderr.
	out := newLineLogger(logger, "[INFO] "+name+" | ")
	cmd.Stdout = out
	cmd.Stderr = out
	// Once stopped, or once exited while a background process still holds its output,
	// do not wait more than that: it would block the whole queue.
	cmd.WaitDelay = waitDelay

	start := time.Now()
	err := NativeExecuteCommand(cmd)
	out.Flush()
	elapsed := time.Since(start).Round(100 * time.Millisecond)
	if ctx.Err() != nil {
		return fmt.Errorf("command %s %v interrupted after %s", name, args, elapsed)
	}
	if errors.Is(err, exec.ErrWaitDelay) {
		logger.Printf("[WARN] %s exited, but left a background process holding its output\n", name)
		err = nil
	}
	if err != nil {
		return fmt.Errorf("command %s %v failed after %s: %w", name, args, elapsed, err)
	}
	logger.Printf("[INFO] %s done in %s\n", name, elapsed)
	return nil
}
