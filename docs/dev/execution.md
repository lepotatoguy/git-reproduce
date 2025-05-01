For building the Go Binary:

For Linux:

`GOOS=linux GOARCH=amd64 go build -o git-reproduce`

For Windows:

`GOOS=windows GOARCH=amd64 go build -o git-reproduce.exe`

For macOS:

`GOOS=darwin GOARCH=amd64 go build -o git-reproduce`

Then, run `./git-reproduce --help`