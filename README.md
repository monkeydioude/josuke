## JoJo's Bizarre webhook handler
_ゴゴゴゴゴゴ ！_

Josuke is a tiny CI/deployment tool that reacts on Gogs/Github/Bitbucket webhook's payload.

Josuke is a simple guy, 3 things and he's happy:
- Write a JSON config file
- Run Josuke and feed him your config
- Go to Github/Bitbucket and set webhooks routes as specified in your config

**Writing a json config file is required.** 

**Config file path must be given using the -c flag** (josuke -c /path/to/config.json).

Example of a classic config.json:

```json
{
    "logLevel": "INFO",
    "host": "127.0.0.1",
    "port": 8082,
    "store": "{directory to store payload, optional)",
    "hook": [
        {
            "name": "gogs",
            "type": "gogs",
            "path": "/josuke/gogs",
            "secret": "7YiuiG8dM1lSh5IzdrVK5XCQcBbRFMvwh5CB4b90"
        },
        {
            "name": "private-gogs",
            "type": "gogs",
            "path": "/josuke/private-gogs",
            "secret": "0061Gki75ieIEWaQ8y8SlGpUhGpx0HEfdF3D61Tz",
            "command": [
                "/home/mkd/Work/josuke/script/hook",
                "%payload_path%"
            ]
        },
        {
            "name": "github",
            "type": "github",
            "path": "/josuke/github",
            "secret": "wd51QvLFIG3VFim5TmltV2xB40YCWwfJmnmxo9pp"
        },
        {
            "name": "bitbucket",
            "type": "bitbucket",
            "path": "/josuke/bitbucket"
        }
    ],
    "deployment":
    [
        {
            "repo": "monkeydioude/donut",
            "proj_dir": "donut",
            "base_dir": "/var/www",
            "branches":
            [
                {
                    "branch" :"master",
                    "actions":
                    [
                        {
                            "action": "push",
                            "commands": [
                                ["echo", "payload written to: ", "%payload_path%"],
                                ["cd", "%base_dir%"],
                                ["git", "clone", "%html_url%"],
                                ["cd", "%proj_dir%"],
                                ["git", "pull", "--rebase"],
                                ["make"]
                            ]
                        }
                    ]
                }
            ]
        }
    ]
}
```

#### TLS configuration ####

Add the `cert` and `key` properties inside config's json file, same level as `host`, `port`. 
```json
{
    "…": "…",
    "port": 8082,
    "cert": "conf/cert.pem",
    "key": "conf/key.pem",
    "…": "…"
}
```

Generate the default certificate and private key with:

```sh
#!/bin/sh
openssl req -x509 -newkey rsa:4096 -nodes \
  -out cert.pem \
  -keyout key.pem -days 365
```

### Keys definition

- `logLevel`: optional, five levels, from the most verbose to the less verbose: `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`. Defaults to `INFO`.
- `host`: binds the server to local address. Defaults to localhost.
- `port`: port Josuke will listen to. Defaults to 8082.
- `store`: directory, optional. If present, every valid payload is written in this directory with a dynamic name: `{hook.name}.{timestamp}.{random string}.json`. The local path to this file is available to commands with the placeholder `%payload_path%`.
- `hook`: array of objects defining SCM hooks for Gogs, GitHub and BitBucket.
- `deployment`: array of objects defining deployments **repository rules** Josuke should follow.

#### Hook Definition ####

- `name` : logical name, used in the payload local file name if enabled.
- `type`: SCM type, currently "gogs", "github" or "bitbucket".
- `path`: local web path. This path must be specified in the SCM webhook’s parameters.
- `secret`: signs the payload for Gogs and Github. *Optional, but strongly recommended for security purpose.* If not set, anybody can fake a payload on your webhook endpoint.
- `command`: optional command, takes precedence over the deployment commands if set. It is run at each valid request. *Only the `%payload_path%` placeholder is available in this hook scope.*  

There are three types of hooks:
- `gogs`
- `github`
- `bitbucket`

##### Command samples #####

Run a shell script `C:/Users/me/josuke/script/hook` with the payload path as a parameter on Windows. It uses the shell that comes with Git Bash:

```json
"command": [
	"C:/Users/me/AppData/Local/Programs/Git/bin/sh.exe",
	"C:/Users/me/josuke/script/hook",
	"%payload_path%"
]
```

#### Repository rules ####

The **repository rules** objects are defined as such:
- `repo`: name of your repository in the repository universe. No need to specify the whole **only the username and repository name is required** (ex: monkeydioude/josuke)
- `branches`: is an array of objects defining the **branche behavior** towards specified branches.
- `base_dir`: **OPTIONAL** Allow you to set what should be a base directory usable at **commands definition** level (ex: /var/projects/sources)
- `proj_dir`: **OPTIONAL** Allow you to set what should be a project directory (or name) usable at **commands definition** level 

**branch behaviors** objects are defined as such:
- `branch`: behavior toward a specific branch
- `actions`: is an array of objects defining the behavior towards specific **actions**.

**actions** objects are defined as such: 
- `action`: is the kind of action sent by the payload, that has been taken toward the source branch (ex: push on a branch, merge a branch with the source branch...)
- `commands`: is an array of objects defining the series of **commands** Josuke should trigger for this `action`

**commands** is an array of array of strings that should contain commands to be executed when an `action` is triggered. 1st index of the array must be the command name. Every following index should be args of the command:
```json
    [
        ["cd", "%base_dir%"],
        ["git", "clone", "%html_url%"],
        ["cd", "%proj_dir%"],
        ["git", "pull", "--rebase"],
        ["make"]
    ]

```

### You can use these 4 placeholders at commands level

- `%base_dir%`: referring to "base_dir" set in config, must be defined by `base_dir` of each `deployment`
- `%proj_dir%`: referring to "proj_dir" set in config, must be defined by `proj_dir` of each `deployment`
- `%html_url%`: retrieved from gogs/github/bitbucket's payload informations, html url of your repo
- `%payload_path%`: path to the payload, available if enabled with `store` in the configuration. Otherwise, empty.

### Tests:

See [testdata/](testdata/index.md).

### Functional tests:
Using `make ftest` will trigger `script/functional-test-runner.sh`. This script will run every script matching `test/functional/test*.sh` pattern.

### Build and run instructions:

__With Golang__:
- Install [the Go language](https://golang.org/dl/)

Then using Makefile (Unix/Linux/MacOS/WSL on Windows):
- `CONF_FILE=/path/to/config/json make go_start`

Or with shell startup script (Unix/Linux/MacOS/WSL on Windows):
- `CONF_FILE=/path/to/config/json script/run.sh`

Or using Golang only (only available option for Windows users not using WSL):
- `go install`
- `josuke -c /path/to/config/json`

__With Docker__
- Install [Docker](https://docs.docker.com/get-docker/)

Then using Makefile (Unix/Linux/MacOS/WSL on Windows):
- `CONF_FILE=/path/to/config/json make start`

Or with Docker only:
- docker build -f build/Dockerfile -t josuke .
- docker run --network="host" -d -e "CONF_FILE=/path/to/config/json" josuke

## Healthcheck:
Once Josuke is running, healthcheck HTTP status is available at `/healthcheck`

## Makefile:
- `install` (dev only): setup dev env such as git hooks
- `stop`: stop josuke running docker container
- `start`: build josuke docker image and run it
- `restart`: `stop` + `start`
- `run`: run a docker container using already built josuke image
- `sr`: `stop` + `run`
- `shell`: run a shell (/bin/sh) in a running josuke container
- `test`: run unit tests inside a container
- `ftest`: run functional tests
- `bb`: rebuild josuke binary inside a running container
- `logs`: read josuke's log file (/var/log/josuke) inside a running container
- `offline_logs`: read logs of the lastet, running or not, josuke container from host's physical log files (/var/lig/docker/containers/$CONTAINER_ID/$CONTAINER_ID-json.log)
- `attach`: attach a tty to a running container. Be advised that detaching the freshly attached tty might require to kill process. ちょっとダメね

Default make rule is `start`

_DORA !_


![](https://68.media.tumblr.com/7b9b18644e2d491cc25267ebde23ec23/tumblr_ohxk9dpmoq1tqvsfso1_540.gif)
