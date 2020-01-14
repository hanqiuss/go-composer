#bin/bash
GOOS=windows GOARCH=amd64 go build -o build/go-composer_win_amd64_beta_0.0.3.exe
GOOS=darwin GOARCH=amd64 go build -o build/go-composer_darwin_amd64_beta-0.0.3
GOOS=linux GOARCH=amd64 go build -o build/go-composer_linux_amd64_beta-0.0.3