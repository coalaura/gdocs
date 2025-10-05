@echo off

echo Building...

setlocal

set "PATH=C:\Windows\system32;C:\Windows;%USERPROFILE%\go\bin;C:\Program Files\Go\bin;C:\Users\Laura\AppData\Local\Microsoft\WinGet\Links"
set PKG_CONFIG=
set CGO_CFLAGS=
set CGO_LDFLAGS=
set CGO_ENABLED=1
set GOOS=linux
set GOARCH=amd64
set CC=zig cc -target x86_64-linux-musl
set CXX=zig c++ -target x86_64-linux-musl

go build -tags musl

endlocal

echo Done.