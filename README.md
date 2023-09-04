# go-authenticator
TOTP password validator written in go language

## Function Introduction
- The program returns 0 indicating successful validation, and 1 indicating failed validation.
- When the program first runs, it will generate 6 emergency codes in your home directory.
- Default timeout is 30 seconds.
- Only one process allowed at a time

## User Interface
![image](https://github.com/loonpn/go-authenticator/assets/107356466/a0dc168d-94f4-49f3-ae43-dab4946a470b)

## Build
```bash
sudo apt install golang
go mod init go-authenticator
go mod tidy
go build
```

## Ussage
### Login Script
Add follow commands to your /home/username/.profile
```bash
# Authenticator
/path/to/go-authenticator
if [ $? -ne 0 ]; then
    logout
fi
```
### ttyd Web Console
Create a shell script:
```bash
#!/bin/sh
/path/to/go-authenticator
if [ $? -eq 0 ]; then
    exec /bin/login
fi
```
Run ttyd using the following command
```bash
ttyd /path/to/shellscript
```
