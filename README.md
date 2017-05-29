**_JoJo's Bizarre webhook handler_**

Josuke is a tiny Github post treatment tool.
After being built and launched on your server, you may write a json config file (default config.json), describing what to do when receiving a payload from Github.
The said config file is an array of json object defined as such:
- Mandatory sections:
        ```
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
                        "commands": []
        ```
- Commands is an optional array of array of strings, it may contains any command you want. 1st index of the array must be the command name. Every following index should be args of the command:
        ```
        [
            ["cd", "%base_dir%"],
            ["git", "clone", "%html_url%"],
            ["cd", "%proj_dir%"],
            ["git", "fetch", "--all"],
            ["git", "checkout", "master"],
            ["git", "reset", "--hard", "origin/master"],
            ["make"]
        ]
        ```

3 Keywords might be used for a lil' dynamic in your deployments:
- %base_dir%: referring to "base_dir" set in config
- %proj_dir%: referring to "proj_dir" set in config
- %html_url%: retrieved from github's payload informations, html url of your repo (might be used, for example, in case of cloning repo)
