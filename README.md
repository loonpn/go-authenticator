# go-authenticator
TOTP password validator written in go language

## Function Introduction
- The program returns 0 indicating successful validation.
- When the program first runs, it will generate 6 emergency codes in your home directory.
- Default timeout is 30 seconds.

## User Interface
![image](https://github.com/loonpn/go-authenticator/assets/107356466/96d81f41-42b1-428e-ac8c-7f7a53f1a282)

## Build
```bash
sudo apt install golang
go mod init go-authenticator
go mod tidy
go build
```

## Ussage
### Login Script
Add follow lines to your /home/username/.profile
```bash
# Authenticator
trap "" INT
if [ -x "/path/to/go-authenticator" ]; then
    /path/to/go-authenticator
    if [ $? -ne 0 ]; then
        logout
    fi
fi
trap - INT
```
### ttyd Web Console
Create a shell script:
```bash
#!/bin/sh
if [ -x "/path/to/go-authenticator" ]; then
    /path/to/go-authenticator
    if [ $? -eq 0 ]; then
        exec /bin/login
    fi
else
    exec /bin/login
fi
```
Run ttyd using the following command
```bash
ttyd /path/to/shellscript
```
