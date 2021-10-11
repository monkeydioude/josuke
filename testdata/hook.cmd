@echo off
@rem Dummy hook, prints the payload path to a log file.
@setlocal
echo processed payload "%1">>"%~dp0hook.log"
@endlocal