## Requirements
1. ffmpeg
2. youtube-dl
3. go >=1.15

## Building
```
$ make
```
or Install
```
$ make install
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

`VERBOSITY` can be any of `(panic|fatal|error|warn|info|debug|trace)` as defined by logrus