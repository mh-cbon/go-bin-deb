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

REPO=`echo ${GH} | cut -d '/' -f 2`
USER=`echo ${GH} | cut -d '/' -f 1`

if ["${GH_TOKEN}" == ""]; then
  echo "GH_TOKEN is not properly set. Check your travis file."
  exit 1
fi

# clean up build.
rm -fr ${REPO}-*.rpm
rm -fr ${REPO}-*.deb

sudo apt-get install build-essential -y

if type "gh-api-cli" > /dev/null; then
  echo "gh-api-cli already installed"
else
  curl -L https://raw.githubusercontent.com/mh-cbon/latest/master/install.sh | GH=mh-cbon/gh-api-cli sh -xe
fi

cd ..
rm -fr ${REPO}
git clone https://github.com/${USER}/${REPO}.git ${REPO}
cd ${REPO}
git config user.name "${USER}"
git config user.email "${EMAIL}"
if [ `git symbolic-ref --short -q HEAD | egrep 'gh-pages$'` ]; then
  echo "already on gh-pages"
else
  if [ `git branch -a | egrep 'remotes/origin/gh-pages$'` ]; then
    # gh-pages already exist on remote
    git checkout gh-pages
  else
    git checkout -b gh-pages
    find . -maxdepth 1 -mindepth 1 -not -name .git -exec rm -rf {} \;
    git commit -am "clean up"
  fi
fi


APTLY="`pwd`/aptly_0.9.7_linux_amd64/aplty"

echo $APTLY 

rm -fr apt
if [ ! -d "aptly_0.9.7_linux_amd64" ]; then
  wget https://bintray.com/artifact/download/smira/aptly/aptly_0.9.7_linux_amd64.tar.gz
  tar xzf aptly_0.9.7_linux_amd64.tar.gz
fi


cat <<EOT > aptly.conf
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

set +x # disable debug output because that would display the token in clear text..
echo "gh-api-cli dl-assets -t {GH_TOKEN} -o ${USER} -r ${REPO} -g '*deb' -out 'pkg/%r-%v_%a.deb'"
gh-api-cli dl-assets -t "${GH_TOKEN}" -o ${USER} -r ${REPO} -g '*deb' -out 'pkg/%r-%v_%a.deb'
set -x

ls -al

if [ ! -d "apt" ]; then
  mkdir apt
  cd apt
  ls -al
  $APTLY repo create -config=../aptly.conf -distribution=all -component=main ${REPO}
  $APTLY repo add -config=../aptly.conf ${REPO} ../pkg
  $APTLY publish -component=contrib -config=../aptly.conf repo ${REPO}
  $APTLY repo show -config=../aptly.conf -with-packages ${REPO}

else
  cd apt
  ls -al
  $APTLY repo add -config=../aptly.conf ${REPO} ../pkg
  $APTLY publish -config=../aptly.conf update all
  $APTLY repo show -config=../aptly.conf -with-packages ${REPO}
fi

cat <<EOT > ${REPO}.list
deb [trusted=yes] https://${USER}.github.io/${REPO}/apt/public/ all contrib
EOT

cd ..
ls -al
rm -f aptly_0.9.7_linux_amd64.tar.gz
rm -f aptly.conf
rm -fr aptly_0.9.7_linux_amd64
rm -fr pkg





git add -A
git commit -m "Created debian repository"

set +x # disable debug output because that would display the token in clear text..
echo "git push --force --quiet https://GH_TOKEN@github.com/${GH}.git gh-pages"
git push --force --quiet "https://${GH_TOKEN}@github.com/${GH}.git" gh-pages \
 2>&1 | sed -re "s/${GH_TOKEN}/GH_TOKEN/g"
