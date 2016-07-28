#!/bin/sh -e

# this is an helper
# to use into your travis file
# it is limited to amd64/386 arch
#
# to use it
# curl -L https://raw.githubusercontent.com/mh-cbon/go-bin-deb/master/create-pkg.sh \
# | GH=mh-cbon/gh-api-cli sh -xe

REPO=`echo ${GH} | cut -d '/' -f 2`
USER=`echo ${GH} | cut -d '/' -f 1`

sudo apt-get install build-essential lintian -y

if type "changelog" > /dev/null; then
  echo "changelog already installed"
else
  curl -L https://raw.githubusercontent.com/mh-cbon/latest/master/install.sh | GH=mh-cbon/changelog sh -xe
fi

if type "go-bin-deb" > /dev/null; then
  echo "go-bin-deb already installed"
else
  curl -L https://raw.githubusercontent.com/mh-cbon/latest/master/install.sh | GH=mh-cbon/go-bin-deb sh -xe
fi

rm -fr pkg-build*/
mkdir -p pkg-build/{386,amd64}

go-bin-deb generate -a 386 --version ${TRAVIS_TAG} -w pkg-build/386/ -o ${TRAVIS_BUILD_DIR}/${REPO}-386.deb
go-bin-deb generate -a amd64 --version ${TRAVIS_TAG} -w pkg-build/amd64/ -o ${TRAVIS_BUILD_DIR}/${REPO}-amd64.deb
