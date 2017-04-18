# go-bin-deb

[![MIT License](http://img.shields.io/badge/License-MIT-yellow.svg)](../LICENSE)

Package go-bin-deb creates binary package for debian system


# Usage

```sh
export GH_TOKEN=`gh-api-cli get-auth -n release`

vagrant up cli

vagrant rsync cli && vagrant ssh cli -c "export GH_TOKEN=$GH_TOKEN; sh /vagrant/vagrant-run.sh"

vagrant destroy cli -f
```

# See also

https://github.com/mh-cbon/go-bin-deb#recipes
