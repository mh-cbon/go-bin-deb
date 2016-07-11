REPO="go-bin-deb"
ARCH=$(uname -m)
case $ARCH in
	arm*) ARCH="arm";;
	x86) ARCH="386";;
	x86_64) ARCH="amd64";;
esac
latest=`wget -q --no-check-certificate -O - https://api.github.com/repos/mh-cbon/${REPO}/releases/latest | grep -E '"tag_name": "([^"]+)"' | cut -d '"' -f4`
wget --no-check-certificate https://github.com/mh-cbon/${REPO}/releases/download/${latest}/${REPO}-${ARCH}.deb
sudo dpkg -i ${REPO}-${ARCH}.deb
