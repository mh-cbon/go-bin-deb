#!/bin/sh -e

# this is an helper
# to use into your travis file
# it is limited to amd64/386 arch
#
# to use it
# curl -L https://raw.githubusercontent.com/mh-cbon/go-bin-deb/master/setup-deb-repository.sh \
# | GH=mh-cbon/gh-api-cli EMAIL=mh-cbon@users.noreply.github.com sh -xe

# GH=$1
# EMAIL=$2

if ["${GH_TOKEN}" == ""]; then
  echo "GH_TOKEN is not properly set. Check your travis file."
  exit 1
fi

if ["${GH}" == ""]; then
  echo "GH is not properly set. Check your travis file."
  exit 1
fi

REPO=`echo ${GH} | cut -d '/' -f 2`
USER=`echo ${GH} | cut -d '/' -f 1`

REPOPATH=`pwd`

# clean up build.
rm -fr ${REPO}-*.rpm
rm -fr ${REPO}-*.deb

# prepare the machine
sudo apt-get install build-essential -y

# install gh-api-cli to dld assets
if type "gh-api-cli" > /dev/null; then
  echo "gh-api-cli already installed"
else
  curl -L https://raw.githubusercontent.com/mh-cbon/latest/master/install.sh | GH=mh-cbon/gh-api-cli sh -xe
fi


cd ${REPOPATH}/..
DREPOPATH="${REPOPATH}/D/"
rm -fr ${DREPOPATH}

# remove exisitng repo

# clone it again
git clone https://github.com/${USER}/${REPO}.git ${DREPOPATH}

# move into, configure git
cd ${DREPOPATH}

git config user.name "${USER}"
git config user.email "${EMAIL}"

git checkout gh-pages | echo "not remote gh pages"

# prepare aptly to generate an apt repo
APTLYDIR="`pwd`/aptly_0.9.7_linux_amd64"
APTLY="`pwd`/aptly_0.9.7_linux_amd64/aptly"
APTLYCONF="${DREPOPATH}/aptly.conf"

# clean up first
rm -fr apt
if [ ! -d "aptly_0.9.7_linux_amd64" ]; then
  wget https://bintray.com/artifact/download/smira/aptly/aptly_0.9.7_linux_amd64.tar.gz
  tar xzf aptly_0.9.7_linux_amd64.tar.gz
fi

# make an aptly.conf
cat <<EOT > ${APTLYCONF}
{
  "rootDir": "`pwd`/apt",
  "downloadConcurrency": 4,
  "downloadSpeedLimit": 0,
  "architectures": [],
  "dependencyFollowSuggests": false,
  "dependencyFollowRecommends": false,
  "dependencyFollowAllVariants": false,
  "dependencyFollowSource": false,
  "gpgDisableSign": true,
  "gpgDisableVerify": true,
  "downloadSourcePackages": false,
  "ppaDistributorID": "",
  "ppaCodename": ""
}
EOT

# dld assets to put in the new repo
set +x # disable debug output because that would display the token in clear text..
echo "gh-api-cli dl-assets -t {GH_TOKEN} -o ${USER} -r ${REPO} -g '*deb' -out 'pkg/%r-%v_%a.deb'"
gh-api-cli dl-assets -t "${GH_TOKEN}" -o ${USER} -r ${REPO} -g '*deb' -out 'pkg/%r-%v_%a.deb'
set -x

# execute aptly
if [ ! -d "apt" ]; then
  mkdir apt
  cd apt
  $APTLY repo create -config=${APTLYCONF} -distribution=all -component=main ${REPO}
  $APTLY repo add -config=${APTLYCONF} ${REPO} ../pkg
  $APTLY publish -component=contrib -config=${APTLYCONF} repo ${REPO}
  $APTLY repo show -config=${APTLYCONF} -with-packages ${REPO}

else
  cd apt
  $APTLY repo add -config=${APTLYCONF} ${REPO} ../pkg
  $APTLY publish -config=${APTLYCONF} update all
  $APTLY repo show -config=${APTLYCONF} -with-packages ${REPO}
fi

# finalize the repo
cat <<EOT > ${REPO}.list
deb [trusted=yes] https://${USER}.github.io/${REPO}/apt/public/ all contrib
EOT

# clean up.
cd ..
rm -rf "${APTLYCONF}" "${APTLY}" "${APTLYDIR}"
rm -rf "${APTLYDIR}/../aptly_0.9.7_linux_amd64.tar.gz"
rm -rf "${APTLYDIR}/../pkg"



git add -A
git commit -m "Created debian repository"

set +x # disable debug output because that would display the token in clear text..
echo "git push --force --quiet https://GH_TOKEN@github.com/${GH}.git gh-pages"
git push --force --quiet "https://${GH_TOKEN}@github.com/${GH}.git" gh-pages \
 2>&1 | sed -re "s/${GH_TOKEN}/GH_TOKEN/g"
