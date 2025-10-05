@echo off

echo Building...

set GOOS=linux
go build -o gdocs

set GOOS=windows

echo Done.