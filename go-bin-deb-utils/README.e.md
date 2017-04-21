---
License: MIT
LicenseFile: ../LICENSE
LicenseColor: yellow
---
# {{.Name}}

{{template "license/shields" .}}

{{pkgdoc}}

# Usage

```sh
export GH_TOKEN=`gh-api-cli get-auth -n release`

vagrant up

vagrant rsync && vagrant ssh -c "export GH_TOKEN=$GH_TOKEN; sh /vagrant/vagrant-run.sh"

vagrant ssh -c 'curl -L https://raw.githubusercontent.com/mh-cbon/latest/master/install.sh | GH=mh-cbon/go-bin-deb sh -xe'
vagrant ssh -c 'go-bin-deb -v'
vagrant ssh -c 'which go-bin-deb'
vagrant ssh -c 'sudo apt-get remove go-bin-deb -y'
vagrant ssh -c 'curl -L https://raw.githubusercontent.com/mh-cbon/latest/master/source.sh | GH=mh-cbon/go-bin-deb sh -xe'
vagrant ssh -c 'go-bin-deb -v'
vagrant ssh -c 'which go-bin-deb'
vagrant ssh -c 'sudo apt-get remove go-bin-deb -y'

vagrant rsync && vagrant ssh -c 'GH=mh-cbon/go-bin-deb sh /vagrant/ppa-add.sh'
vagrant rsync && vagrant ssh -c 'sudo apt-cache showpkg go-bin-deb'
vagrant rsync && vagrant ssh -c 'sudo apt-cache showpkg go-bin-rpm'
vagrant rsync && vagrant ssh -c 'sudo apt-get remove go-bin-deb -y'
vagrant rsync && vagrant ssh -c 'sudo apt-get install go-bin-deb -y --force-yes'
vagrant ssh -c 'go-bin-deb -v'
vagrant rsync && vagrant ssh -c 'sudo apt-get install go-bin-deb=0.0.17 --force-yes -y'
vagrant rsync && vagrant ssh -c 'sudo apt-get update'
vagrant ssh -c 'go-bin-deb -v'
vagrant rsync && vagrant ssh -c 'sudo apt-mark hold go-bin-deb'


vagrant destroy -f
```

# See also

https://askubuntu.com/a/138327/544699
https://github.com/mh-cbon/go-bin-deb#recipes
