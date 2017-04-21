#!/bin/sh -e

set -e
set -x

export GOINSTALL="/go"
export GOROOT=${GOINSTALL}/go/
export PATH=$PATH:$GOROOT/bin

getgo="https://raw.githubusercontent.com/mh-cbon/latest/master/get-go.sh?d=`date +%F_%T`"
# install go, specific to vagrant
if type "wget" > /dev/null; then
  wget --quiet -O - $getgo | sh -xe
fi
if type "curl" > /dev/null; then
  curl -s -L $getgo | sh -xe
fi

echo "$PATH"
go version
go env

export GOPATH=/gopath/
export PATH=$PATH:/gopath/bin

sudo chown -R vagrant:vagrant -R $GOPATH
mkdir -p ${GOPATH}/bin

[ -d "$GOPATH" ] || echo "$GOPATH does not exists"
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

cd $GOPATH/src/github.com/$REPO/go-bin-deb-utils
go install

cd $GOPATH/src/github.com/$REPO
[ -d "$GOPATH/src/github.com/$REPO/vendor" ] || glide install
go build

# build the binaries
BINBUILD_DIR="$GOPATH/src/github.com/$REPO/build"
rm -fr "$BINBUILD_DIR"
mkdir -p "$BINBUILD_DIR/{386,amd64}"

PKGBUILD_DIR="$GOPATH/src/github.com/$REPO/apt"
PPABUILD_DIR="$GOPATH/src/github.com/$REPO/ppa"

# rm -fr $PKGBUILD_DIR $PPABUILD_DIR

# build the packages
set +x
echo ""
echo "# =================================================="
echo "# =================================================="
set -x
f="-X main.VERSION=${VERSION}"
GOOS=linux GOARCH=386 go build --ldflags "$f" -o "$BINBUILD_DIR/386/go-bin-deb" $k
GOOS=linux GOARCH=amd64 go build --ldflags "$f" -o "$BINBUILD_DIR/amd64/go-bin-deb" $k
go-bin-deb-utils create-packages -repo=$REPO

set +x
echo ""
echo "# =================================================="
echo "# =================================================="
set -x
go-bin-deb-utils setup-repository -out="${PKGBUILD_DIR}" -push -repo=$REPO

set +x
echo ""
echo "# =================================================="
echo "# =================================================="
set -x
go-bin-deb-utils setup-ppa -out="${PPABUILD_DIR}" -push -repo=$REPO -repos="mh-cbon/go-bin-deb,mh-cbon/go-bin-rpm"

set +x
echo ""
echo "# =================================================="
echo "# =================================================="
echo "      All Done!"
set -x

#
