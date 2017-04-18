#!/bin/sh -e

# this is an helper
# to use into your travis file
# golang is required.
#
# to use it
# curl -L https://raw.githubusercontent.com/mh-cbon/go-bin-deb/master/create-pkg.sh \
# | GH=YOUR/REPO sh -xe

if [ "${GH}" = "mh-cbon/go-bin-deb" ]; then
  git remote -vv
  git branc -aav
  git fetch origin
  git checkout -b master
  git pull origin
  ls -al
  go run go-bin-deb-utils/*go create-packages -repo=$GH
else
  go get -u github.com/mh-cbon/go-bin-deb/go-bin-deb-utils
  go-bin-deb-utils create-packages -repo=$GH
fi
