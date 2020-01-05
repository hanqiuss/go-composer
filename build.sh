#bin/bash
GOOS=windows GOARCH=amd64 go build -o go-composer_win_amd64.exe
GOOS=darwin GOARCH=amd64 go build -o go-composer_darwin_amd64
GOOS=linux GOARCH=amd64 go build -o go-composer_linux_amd64