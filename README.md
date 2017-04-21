emd gen -in README.e.md
# go-bin-deb

[![travis Status](https://travis-ci.org/mh-cbon/go-bin-deb.svg?branch=master)](https://travis-ci.org/mh-cbon/go-bin-deb) [![Go Report Card](https://goreportcard.com/badge/github.com/mh-cbon/go-bin-deb)](https://goreportcard.com/report/github.com/mh-cbon/go-bin-deb) [![GoDoc](https://godoc.org/github.com/mh-cbon/go-bin-deb?status.svg)](http://godoc.org/github.com/mh-cbon/go-bin-deb) [![MIT License](http://img.shields.io/badge/License-MIT-yellow.svg)](../LICENSE)

Package go-bin-deb creates binary package for debian system


Using a `json` files to declare rules,
it then performs necessary operations to invoke `dpkg-deb` to build a package,
then check it with the help of `lintian`.

This tool is part of the [go-github-release workflow](https://github.com/mh-cbon/go-github-release)

See [the demo](demo/).

# TOC
- [Install](#install)
  - [Glide](#glide)
  - [linux rpm/deb repository](#linux-rpmdeb-repository)
  - [linux rpm/deb standalone package](#linux-rpmdeb-standalone-package)
- [Requirements](#requirements)
- [Usage](#usage)
  - [Workflow overview](#workflow-overview)
  - [Json file](#json-file)
- [CLI](#cli)
  - [generate](#generate)
  - [test](#test)
- [Recipes](#recipes)
  - [Installing generated package](#installing-generated-package)
  - [Vagrant recipe](#vagrant-recipe)
  - [Travis recipe](#travis-recipe)
  - [useful deb commands](#useful-deb-commands)
  - [Release the project](#release-the-project)
- [History](#history)

# Install

Check the [release page](https://github.com/mh-cbon/go-bin-deb/releases)!

#### Glide
```sh
mkdir -p $GOPATH/src/github.com/mh-cbon/go-bin-deb
cd $GOPATH/src/github.com/mh-cbon/go-bin-deb
git clone https://github.com/mh-cbon/go-bin-deb.git .
glide install
go install
```

#### linux rpm/deb repository
```sh
wget -O - https://raw.githubusercontent.com/mh-cbon/latest/master/source.sh \
| GH=mh-cbon/go-bin-deb sh -xe
# or
curl -L https://raw.githubusercontent.com/mh-cbon/latest/master/source.sh \
| GH=mh-cbon/go-bin-deb sh -xe
```

#### linux rpm/deb standalone package
```sh
curl -L https://raw.githubusercontent.com/mh-cbon/latest/master/install.sh \
| GH=mh-cbon/go-bin-deb sh -xe
# or
wget -q -O - --no-check-certificate \
https://raw.githubusercontent.com/mh-cbon/latest/master/install.sh \
| GH=mh-cbon/go-bin-deb sh -xe
```

# Requirements

A debian system, vagrant, travis, docker, whatever.

# Usage

## Workflow overview

To create a binary package you need to

- build your application binaries
- invoke `go-bin-deb` to generate the package
- create deb repositories on `travis` hosted on `gh-pages` using this [script](setup-repository.sh)

## Json file

For a reference of all fields, see [this](deb-example.json)

For a real world example including service, shortcuts, env, see [this](demo/deb.json)

For a casual example to provide a simple binary, see [this](deb.json)

# CLI
go-bin-deb -help
```sh
NAME:
   go-bin-deb - Generate a binary debian package

USAGE:
   go-bin-deb <cmd> <options>
   
VERSION:
   0.0.0
   
COMMANDS:
     generate  Generate the contents of the package
     test      Test the package json file
     help, h   Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

### generate
go-bin-deb generate -help
```sh
NAME:
   go-bin-deb generate - Generate the contents of the package

USAGE:
   go-bin-deb generate [command options] [arguments...]

OPTIONS:
   --wd value, -w value      Working directory to prepare the package (default: "pkg-build")
   --output value, -o value  Output directory for the debian package files
   --file value, -f value    Path to the deb.json file (default: "deb.json")
   --version value           Version of the package
   --arch value, -a value    Arch of the package
```

### test
go-bin-deb test -help
```sh
NAME:
   go-bin-deb test - Test the package json file

USAGE:
   go-bin-deb test [command options] [arguments...]

OPTIONS:
   --file value, -f value  Path to the deb.json file (default: "deb.json")
```

# Recipes

### Installing generated package

__TLDR__

```sh
# install a package with dependencies
dpkg -i mypackage.deb
apt-get install --fix-missing
# or
gdebi mypackage.deb
```

On debian system to install a package `.deb` file, you should use `dpkg -i` and not `apt-get i`.

But, `dpkg` does not install dependencies by itself, thus you will need to execute an extra command
`apt-get i --fix-missing` to locate and install missing dependencies ater you installed your own `.deb`.

An alternative is to use `gdebi`, which appears to be bale to all of that in one command.

Finally, if one provides a web interface to host the package, it should be no problem to use a regular `apt-get`.

PS: To remove the package `dpkg -r`.

### Vagrant recipe

Please check the demo app [here](demo/)

### Travis recipe

- get a github repo
- create a token in your settings
- get a travis account
- connect your github account to travis and register your repo
- install travis client `gem install --user travis`
- run `travis encrypt --add -r your/repo GH_TOKEN=xxxx`
- run `travis setup releases`
- personalize the `.travis.yml`

```yml
sudo: required

services:
  - docker

language: go
go:
  - tip

env:
  global:
    - MYAPP=dummy
    - MYEMAIL=some@email.com
    - secure: GH_TOKEN xxxx

before_install:
  - sudo apt-get -qq update
  - mkdir -p ${GOPATH}/bin

install:
  - cd $GOPATH/src/github.com/YOUR_USERNAME/$MYAPP
  - go install

script: echo "pass"

# build the app, build the package
before_deploy:
  - mkdir -p build/{386,amd64}
  - GOOS=linux GOARCH=386 go build --ldflags "-X main.VERSION=${TRAVIS_TAG}" -o build/386/$MYAPP main.go
  - GOOS=linux GOARCH=amd64 go build --ldflags "-X main.VERSION=${TRAVIS_TAG}" -o build/amd64/$MYAPP main.go
  - curl -L https://raw.githubusercontent.com/mh-cbon/go-bin-deb/master/create-pkg.sh | GH=YOUR_USERNAME/$MYAPP sh -xe

# build the package repo on gh-pages
after_deploy:
  - curl -L https://raw.githubusercontent.com/mh-cbon/go-bin-deb/master/setup-repository.sh | GH=YOUR_USERNAME/$MYAPP EMAIL=$MYEMAIL sh -xe

# deploy package files into gh releases
deploy:
  provider: releases
  api_key:
    secure: GH_TOKEN xxxx
  file_glob: true
  file:
    - $MYAPP-386.deb
    - $MYAPP-amd64.deb
  skip_cleanup: true
  on:
    tags: true
```

### useful deb commands

```sh
# Install required dependencies to build a package
sudo apt-get install build-essential lintian -y
# build a bin package
dpkg-deb --build debian hello.deb
# show info of a package
dpkg-deb --show hello.deb
# list contents of a package
dpkg-deb --contents hello.deb
```

### Release the project

```sh
gump patch -d # check
gump patch # bump
```

# History

[CHANGELOG](CHANGELOG.md)
[32m ✔ Success[0m
