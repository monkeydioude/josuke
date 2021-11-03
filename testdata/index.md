Test
====

## Configuration ##

[test2.config.json](test2.config.json)

## Scripts ##

- [`send-payload`](send-payload): usage `./send-payload <server JSON conf> <hook name> <payload path>`.
- [`test`](test) : usage: `./test <hook name>`. It calls `send-payload` with derived options.

## Test ##

1. Edit the configuration `testadata/test2.config.json`.
2. Start the server `josuke -c testadata/test2.config.json`
3. Send the [default test payload](commit-payload.json) with `./test`.
4. Send the [gogs test payload](gogs-push.json) with `./test gogs`.
