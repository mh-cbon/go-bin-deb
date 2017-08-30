#!/bin/sh -e

# deprecated script.

echo "You are using a deprecated script. please update your build with latest changes on go-github-release"

# this is an helper
# to use into your travis file
# golang is required.
#
# to use it
# curl -L https://raw.githubusercontent.com/mh-cbon/go-bin-deb/master/setup-repository.sh \
# | GH=YOUR/REPO sh -xe

go get -u github.com/mh-cbon/go-bin-deb/go-bin-deb-utils
go-bin-deb-utils setup-repository -out="`pwd`/apt" -push -repo=$GH


# deprecated script

echo "You are using a deprecated script. please update your build with latest changes on go-github-release"
