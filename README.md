## Requirements
1. ffmpeg
2. youtube-dl
3. go >=1.25
4. protoc, protoc-gen-go, protoc-gen-go-grpc

## Building
```
$ make
```
or Install
```
$ make install
```

## Pre-Requisites
### Install Protoc and Protogen
```sh
URL=$(curl -s https://api.github.com/repos/protocolbuffers/protobuf/releases/latest \
  | jq -r '.assets[] | select(.name | endswith("linux-x86_64.zip")) | .browser_download_url')
curl -LO $URL
unzip $(basename $URL) -d $HOME/.local
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

### Install ffmpeg
```
sudo apt-get install ffmpeg
sudo apt-get install libopus-dev
```

## Running
Use flags
```sh
$ out/costanza -v VERBOSITY -t DISCORD_TOKEN --key GOOGLE_API_KEY
```
or Environment Variables
```sh
$ GOOGLE_KEY=KEY COSTANZA_TOKEN=TOKEN out/costanza -v VERBOSITY
```

or with .env
```sh
set -a; source .env; out/costanza -v VERBOSITY
```


`VERBOSITY` can be any of `(panic|fatal|error|warn|info|debug|trace)` as defined by logrus
