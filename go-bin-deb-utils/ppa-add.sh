#!/bin/sh -e

# deprecated script.

echo "You are using a deprecated script. please update your build with latest changes on go-github-release"

# this is an helper
# to install a source repo
# hosted on gh-pages
#
# to use it
# curl -L https://raw.githubusercontent.com/mh-cbon/latest/master/source.sh \
# | GH=mh-cbon/gh-api-cli sh -xe

# GH=$1

set -x

REPO=`echo ${GH} | cut -d '/' -f 2`
USER=`echo ${GH} | cut -d '/' -f 1`

if "${GH}"=""; then
  exit "GH is required"
fi

DLCMD=""
DLArgs=""
FILE=""
URL=""

if type "dpkg" > /dev/null; then
  FILE=/etc/apt/sources.list.d/${REPO}.list
  URL=http://${USER}.github.io/${REPO}/ppa/${REPO}.list
elif type "dnf" > /dev/null; then
  FILE=/etc/yum.repos.d/${REPO}.repo
  URL=http://${USER}.github.io/${REPO}/rpm/${REPO}.repo
elif type "yum" > /dev/null; then
  FILE=/etc/yum.repos.d/${REPO}.repo
  URL=http://${USER}.github.io/${REPO}/rpm/${REPO}.repo
fi

if type "wget" > /dev/null; then
  DLCMD='wget -q -O '
  DLArgs="${FILE} ${URL}"
elif type "curl" > /dev/null; then
  DLCMD='curl -s -L'
  DLArgs="${URL} > ${FILE}"
fi

if type "sudo" > /dev/null; then
  sudo /bin/sh -c "${DLCMD} ${DLArgs}"
else
  $DLCMD ${DLArgs}
fi


if type "dpkg" > /dev/null; then
  if type "sudo" > /dev/null; then
    sudo apt-get install apt-transport-https -y --quiet
    sudo apt-get update --quiet
    sudo apt-get install ${REPO} -y --quiet
  else
    apt-get install apt-transport-https -y --quiet
    apt-get update --quiet
    apt-get install ${REPO} -y --quiet
  fi

elif type "dnf" > /dev/null; then
  if type "sudo" > /dev/null; then
    sudo dnf install ${REPO} -y --quiet
  else
    dnf install ${REPO} -y --quiet
  fi

elif type "yum" > /dev/null; then
  if type "sudo" > /dev/null; then
    sudo yum install ${REPO} -y --quiet
  else
    yum install ${REPO} -y --quiet
  fi

fi
