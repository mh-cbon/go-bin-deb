
set -e
set -x


rm -fr build && mkdir -p build/{386,amd64}
GOOS=linux GOARCH=386 go build -o build/386/hello hello.go
GOOS=linux GOARCH=amd64 go build -o build/amd64/hello hello.go
go build -o go-bin-deb ../*.go
vagrant up
vagrant rsync
vagrant ssh -c 'sh /vagrant/test-pkg.sh' || "echo keep going"
vagrant ssh -c "sudo apt-get install software-properties-common -y"
vagrant ssh -c "sudo add-apt-repository 'deb https://dl.bintray.com/mh-cbon/deb unstable main'"
vagrant ssh -c "sudo apt-get -qq update"
vagrant ssh -c "sudo apt-get install --allow-unauthenticated -y changelog go-bin-deb emd go-msi gump gh-api-cli"
vagrant ssh -c "changelog -v && gh-api-cli -v && go-bin-deb -v && go-msi -v && gump -v && emd -version"
rm ./go-bin-deb
