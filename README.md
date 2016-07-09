
```sh
vagrant ssh -c "sudo apt-get install build-essential lintian devscripts -y"
vagrant ssh -c "cd /vagrant && rm -f hello.deb"
vagrant ssh -c "cd /vagrant/pkg-build && dpkg-deb --build debian hello.deb"
vagrant ssh -c "cd /vagrant && dpkg-deb --show hello.deb"
vagrant ssh -c "cd /vagrant && dpkg-deb --contents hello.deb"

go build -o go-bin-deb main.go && vagrant rsync && vagrant ssh -c "cd /vagrant && VERBOSE=* ./go-bin-deb generate && cd pkg-build && dpkg-deb --build debian hello.deb && dpkg-deb --contents hello.deb"
```
