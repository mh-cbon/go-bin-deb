
set -e
set -x


rm -fr build && mkdir -p build/{386,amd64}
GOOS=linux GOARCH=386 go build -o build/386/hello hello.go
GOOS=linux GOARCH=amd64 go build -o build/amd64/hello hello.go
go build -o go-bin-deb ../*.go
vagrant up
vagrant rsync
vagrant ssh -c 'sh /vagrant/test-pkg.sh'
rm go-bin-deb
