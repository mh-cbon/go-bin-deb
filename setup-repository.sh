#!/bin/sh -e

# this is an helper
# to use into your travis file
# golang is required.
#
# to use it
# curl -L https://raw.githubusercontent.com/mh-cbon/go-bin-deb/master/setup-deb-repository.sh \
# | GH=YOUR/REPO sh -xe

go get -u github.com/mh-cbon/go-bin-deb/go-bin-deb-utils
go-bin-deb-utils setup-repository -push -repo=$REPO
