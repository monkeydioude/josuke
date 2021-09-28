## JoJo's Bizarre webhook handler
_ゴゴゴゴゴゴ ！_

Josuke is a tiny Github post treatment tool.

**Writing a json config file is required.** 

**Config file path must be given using the -c flag** (josuke -c /path/to/config.json).

Example of a classic config.json:

```json
{
    "github_hook": "/josuke/github",
    "bitbucket_hook": "/josuke/bitbucket",
    "host": "127.0.0.1",
    "port": 8082,
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

Add the `cert` and `key` properties inside config's json file, same level as `github_hook`, `bitbucket_hook` etc... 
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
- `github_hook`: route Josuke will be receiving Github's payload. **Must be specified in Github Webhooks' parameters**
- `bitbucket_hook`: route Josuke will be receiving Bitbuckets's payload. **Must be specified in Bitbuckets Webhooks' parameters**
- `port`: port Josuke will listen to
- `deployment`: array of objects defining deployments **repository rules** Josuke should follow.

These **repository rules** objects are defined as such:
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

### You can use these 3 Keywords at commands level
- `%base_dir%`: referring to "base_dir" set in config, must be defined by `base_dir` of each `deployment`
- `%proj_dir%`: referring to "proj_dir" set in config, must be defined by `proj_dir` of each `deployment`
- `%html_url%`: retrieved from github/bitbucket's payload informations, html url of your repo


### Incoming:
- Tests
- Docker image for testing/building
- Makefile for all of this
- Go1.11 Module compliancy

_DORA !_



![](https://68.media.tumblr.com/7b9b18644e2d491cc25267ebde23ec23/tumblr_ohxk9dpmoq1tqvsfso1_540.gif)
