# hello package - a demo

`hello` is a program which serves a web server on port 8080.

The package installs
- the program bin and its assets
- a service unit file for systemd
- a desktop link to open the hello homepage
- and environment variable

# run it

```sh
rm -fr build && mkdir -p build/{386,amd64}
GOOS=linux GOARCH=386 go build -o build/386/hello hello.go
GOOS=linux GOARCH=amd64 go build -o build/amd64/hello hello.go
vagrant up
vagrant ssh -c 'sudo apt-get install build-essential lintian -y'
vagrant ssh -c 'wget -q -O - --no-check-certificate https://raw.githubusercontent.com/mh-cbon/go-bin-deb/master/install.sh | sh'
vagrant ssh -c 'cd /vagrant/ && VERBOSE=* go-bin-deb generate -a 386 --version 0.0.1 -w pkg-build/386/ -o hello-386.deb'
vagrant ssh -c 'cd /vagrant/ && VERBOSE=* go-bin-deb generate -a amd64 --version 0.0.1 -w pkg-build/amd64/ -o hello-amd64.deb'
vagrant ssh -c 'cd /vagrant/ && dpkg-deb --show hello-amd64.deb'
vagrant ssh -c 'cd /vagrant/ && dpkg-deb --contents hello-amd64.deb'
vagrant ssh -c "sudo dpkg -r hello"
vagrant ssh -c "cd /vagrant/ && sudo dpkg -i hello-amd64.deb"
vagrant ssh -c "echo \$some"
vagrant ssh -c "hello -h"
vagrant ssh -c "systemctl status hello"
vagrant ssh -c "wget -q -O - http://localhost:8080/"
```
