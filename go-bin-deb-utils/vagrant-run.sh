#!/bin/sh -e

set -e
set -x

# install go, specific to vagrant
if type "go" > /dev/null; then
  echo "go already installed"
else
  sudo mkdir -p /go/
  sudo chown -R vagrant:vagrant -R /go
  cd /go/
  [ -f "go1.8.1.linux-amd64.tar.gz" ] || wget https://storage.googleapis.com/golang/go1.8.1.linux-amd64.tar.gz
  [ -d "go" ] || tar -xf go1.8.1.linux-amd64.tar.gz
  ls -al .
  ls -al ./go/bin
  export GOROOT=/go/go/
  export PATH=$PATH:$GOROOT/bin
  cd ~
fi

export GOROOT=/go/go/
export PATH=$PATH:$GOROOT/bin

go version
go env

export GOPATH=/gopath/
export PATH=$PATH:/gopath/bin

sudo chown -R vagrant:vagrant -R $GOPATH
mkdir -p ${GOPATH}/bin

[ -d "$GOPATH" ] || echo "$GOPATH does not exists, do you run vagrant ?"
[ -d "$GOPATH" ] || exit 1;


set +x
# everything here will be replicated into the CI build file (.travis.yml)
export GH_TOKEN="$GH_TOKEN"
NAME="go-bin-deb"
export REPO="mh-cbon/$NAME"
export EMAIL="mh-cbon@users.noreply.github.com"

# set env specific to vagrant
export VERSION="LAST"
export BUILD_DIR="$GOPATH/src/github.com/$REPO/pkg-build"

# set env specific to travis
# export TRAVIS_TAG="0.0.1-beta999"
# export TRAVIS_BUILD_DIR="$GOPATH/src/github.com/$REPO/pkg-build"

set -x
# setup glide
if type "glide" > /dev/null; then
  echo "glide already installed"
  glide -v
else
  curl https://glide.sh/get | sh
fi

cd $GOPATH/src/github.com/$REPO
git checkout master
git reset HEAD --hard
[ -d "$GOPATH/src/github.com/$REPO/vendor" ] || glide install
go install

# build the binaries
BINBUILD_DIR="$GOPATH/src/github.com/$REPO/build"
rm -fr "$BINBUILD_DIR"
mkdir -p "$BINBUILD_DIR/{386,amd64}"

f="-X main.VERSION=${VERSION}"
GOOS=linux GOARCH=386 go build --ldflags "$f" -o "$BINBUILD_DIR/386/go-bin-deb" $k
GOOS=linux GOARCH=amd64 go build --ldflags "$f" -o "$BINBUILD_DIR/amd64/go-bin-deb" $k

# build the packages
go run /vagrant/*go create-packages -push -repo=$REPO
go run /vagrant/*go setup-repository -push -repo=$REPO

#
