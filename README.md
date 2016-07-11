# go-bin-deb

Create binary package for debian system, see also [the demo](demo/).

Using a `json` files to declare rules, it then performs necessary operations to invoke `dpkg-deb` to build the package, then check it with the help of `lintian`.

__wip__

# Install

```sh
mkdir -p $GOPATH/github.com/mh-cbon
cd $GOPATH/github.com/mh-cbon
git clone https://github.com/mh-cbon/go-bin-deb.git
cd go-bin-deb
glide install
go install
```

# Requirements

A debian system, vagrant, travis, docker, whatever.

# Usage

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

#### generate

```sh
NAME:
   main generate - Generate the contents of the package

USAGE:
   main generate [command options] [arguments...]

OPTIONS:
   --wd value, -w value      Working directory to prepare the package (default: "pkg-build")
   --output value, -o value  Output directory for the debian package files
   --file value, -f value    Path to the deb.json file (default: "deb.json")
   --version value           Version of the package
   --arch value, -a value    Arch of the package
```

#### test

```sh
NAME:
   main test - Test the package json file

USAGE:
   main test [command options] [arguments...]

OPTIONS:
   --file value, -f value  Path to the deb.json file (default: "deb.json")
```

# Installing generated package

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

# Json file

For a reference of all fields, see [this](deb-example.json)

For a real world example including service, shortcuts, env, see [this](demo/deb.json)

For a casual example to provide a simple binary, see [this](deb.json)

# Travis recipe

__wip__

- get a github repo
- get a travis account
- connect your github account to travis and register your repo
- install travis client `gem install --user travis`
- run `travis setup releases`
- personalize the `.travis.yml`

```yml
language: go
go:
  - tip
before_install:
  - sudo apt-get -qq update
  - sudo apt-get install build-essential lintian -y
  - curl https://glide.sh/get | sh
install:
  - glide install
before_deploy:
  - mkdir -p build/{386,amd64}
  - GOOS=linux GOARCH=386 go build -o build/386/program main.go
  - GOOS=linux GOARCH=amd64 go build -o build/amd64/program main.go
deploy:
  provider: releases
  api_key:
    secure: ... your own here
  file:
    - go-bin-deb-386.deb
    - go-bin-deb-amd64.deb
  skip_cleanup: true
  on:
    tags: true
```

# useful deb commands

```sh
# Install required dependencies to build a package
vagrant ssh -c 'sudo apt-get install build-essential lintian -y'
# build a bin package
vagrant ssh -c "cd /vagrant/pkg-build && dpkg-deb --build debian hello.deb"
# show info of a package
vagrant ssh -c "cd /vagrant && dpkg-deb --show hello.deb"
# list contents of a package
vagrant ssh -c "cd /vagrant && dpkg-deb --contents hello.deb"
```

# Testing

Some commands i used to test / develop this program,

```sh
# remove the previously installed package
vagrant ssh -c "sudo dpkg -r hello"

# generate a new package
go build -o go-bin-deb main.go && vagrant rsync && vagrant ssh -c "cd /vagrant/demo && rm -fr /tmp/test && VERBOSE=* ../go-bin-deb generate -a amd64 --version 0.0.1 -w /tmp/test -o hello-amd64.deb"

# install the package
vagrant ssh -c "sudo dpkg -i /vagrant/demo/hello-amd64.deb"

# inspect the package
vagrant ssh -c "cd /vagrant/demo && ar vx hello-amd64.deb && tar -xvf data.tar.xz && tar -xzvf control.tar.gz && ls -alh"

# same with 386 arch version
go build -o go-bin-deb main.go && vagrant rsync && vagrant ssh -c "cd /vagrant/demo && rm -fr /tmp/test && VERBOSE=* ../go-bin-deb generate -a 386 --version 0.0.1 -w /tmp/test -o hello-386.deb"

# that env var should be installed by the package
vagrant ssh -c "echo \$some"

# the command should work and have appropriate permissions
vagrant ssh -c "hello"
```
