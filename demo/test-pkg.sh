
set -e
set -x
sudo apt-get install build-essential lintian -y
# wget -q -O - --no-check-certificate https://raw.githubusercontent.com/mh-cbon/latest/master/install.sh | GH=mh-cbon/go-bin-deb sh -xe
# generate packages
cd /vagrant/ && VERBOSE=* ./go-bin-deb generate -a 386 --version 0.0.1 -w pkg-build/386/ -o hello-386.deb
cd /vagrant/ && VERBOSE=* ./go-bin-deb generate -a amd64 --version 0.0.1 -w pkg-build/amd64/ -o hello-amd64.deb
# inspect the package
cd /vagrant/ && dpkg-deb --show hello-amd64.deb
cd /vagrant/ && dpkg-deb --contents hello-amd64.deb
# remove package
sudo dpkg -r hello
sudo systemctl daemon-reload
systemctl status hello
# install the package
cd /vagrant/ && sudo dpkg -i hello-amd64.deb
echo \$some
which hello
systemctl status hello
wget -q -O - http://localhost:8080/
