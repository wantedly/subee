![subee](docs/subee.png)

[![CircleCI](https://circleci.com/gh/wantedly/subee/tree/master.svg?style=svg)](https://circleci.com/gh/wantedly/subee/tree/master)
[![latest](https://img.shields.io/github/release/wantedly/subee.svg)](https://github.com/wantedly/subee/releases/latest)
[![GoDoc](https://godoc.org/github.com/wantedly/subee?status.svg)](https://godoc.org/github.com/wantedly/subee)
[![Go Report Card](https://goreportcard.com/badge/github.com/wantedly/subee)](https://goreportcard.com/report/github.com/wantedly/subee)
[![license](https://img.shields.io/github/license/wantedly/subee.svg)](./LICENSE)

## Examples

- [Your first Google Cloud Pub/sub app](_examples/your-first-cloudpubsub-app/) - start here!

## CLI

We offer an optional cli tool to develop with subee faster.

```
$ subee help

Usage:
  subee [command]

Available Commands:
  generate    Generate a new code
  help        Help about any command
  start       Build and start subscribers
  version     Print the version information

Flags:
  -h, --help   help for subee

Use "subee [command] --help" for more information about a command.
```

### Installing

#### macOS

```
brew install wantedly/tap/subee
```

#### Other platforms

You can download prebuilt binaries for each platform in the [releases](https://github.com/wantedly/subee/releases) page.

### Build from source

```
go get github.com/wantedly/subee/cmd/subee
```
