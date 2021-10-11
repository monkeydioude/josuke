#!/bin/sh
# Dummy hook, prints the payload path to a log file.
printf 'processed payload %s\n' "$1" >>hook.log
