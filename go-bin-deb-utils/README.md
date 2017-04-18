# go-bin-deb-utils

[![<no value> License](http://img.shields.io/badge/License-<no value>-blue.svg)](LICENSE)

go-bin-deb-utils is a cli tool to generate debian package and repos.


# Usage

```sh
export GHTOKEN=`gh-api-cli get-auth -n release`

vagrant up cli

vagrant rsync cli && vagrant ssh cli -c "export GHTOKEN=$GHTOKEN; sh /vagrant/vagrant-run.sh"

vagrant destroy cli -f
```

# See also

https://github.com/mh-cbon/go-bin-deb#recipes
