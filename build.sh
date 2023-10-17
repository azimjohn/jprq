GOOS=darwin     GOARCH=amd64    go build -ldflags '-s' -o bin/jprq-darwin-amd64       cli/*.go
GOOS=darwin     GOARCH=arm64    go build -ldflags '-s' -o bin/jprq-darwin-arm64       cli/*.go
GOOS=linux      GOARCH=386      go build -ldflags '-s' -o bin/jprq-linux-386          cli/*.go
GOOS=linux      GOARCH=amd64    go build -ldflags '-s' -o bin/jprq-linux-amd64        cli/*.go
GOOS=linux      GOARCH=arm      go build -ldflags '-s' -o bin/jprq-linux-arm          cli/*.go
GOOS=linux      GOARCH=arm64    go build -ldflags '-s' -o bin/jprq-linux-arm64        cli/*.go
GOOS=windows    GOARCH=386      go build -ldflags '-s' -o bin/jprq-windows-386.exe    cli/*.go
GOOS=windows    GOARCH=amd64    go build -ldflags '-s' -o bin/jprq-windows-amd64.exe  cli/*.go
