
```sh
vagrant ssh -c "sudo apt-get install build-essential lintian devscripts -y"
vagrant ssh -c "cd /vagrant && rm -f hello.deb"
vagrant ssh -c "cd /vagrant/pkg-build && dpkg-deb --build debian hello.deb"
vagrant ssh -c "cd /vagrant && dpkg-deb --show hello.deb"
vagrant ssh -c "cd /vagrant && dpkg-deb --contents hello.deb"

go build -o go-bin-deb main.go && vagrant rsync && vagrant ssh -c "cd /vagrant && VERBOSE=* ./go-bin-deb generate && cd pkg-build && dpkg-deb --build debian hello.deb && dpkg-deb --contents hello.deb"


go build -o go-bin-deb main.go && vagrant rsync && vagrant ssh -c "cd /vagrant/demo && rm -fr /tmp/test && VERBOSE=* ../go-bin-deb generate -a amd64 --version 0.0.1 -w /tmp/test -o hello-amd64.deb"
vagrant ssh -c "sudo dpkg -r hello"
vagrant ssh -c "cd /vagrant/demo && ar vx hello-amd64.deb && tar -xvf data.tar.xz && tar -xzvf control.tar.gz && ls -alh"
vagrant ssh -c "sudo dpkg -i /vagrant/demo/hello-amd64.deb"

go build -o go-bin-deb main.go && vagrant rsync && vagrant ssh -c "cd /vagrant/demo && rm -fr /tmp/test && VERBOSE=* ../go-bin-deb generate -a 386 --version 0.0.1 -w /tmp/test -o hello-386.deb"

vagrant ssh -c "sudo dpkg -i /vagrant/demo/hello-386.deb"
vagrant ssh -c "echo \$some"
vagrant ssh -c "hello"
```

```
dpkg -i mypackage.deb
apt-get install --fix-missing
# or
gdebi mypackage.deb
```
