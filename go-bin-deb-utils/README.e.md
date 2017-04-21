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

vagrant destroy -f
```

# See also

https://github.com/mh-cbon/go-bin-deb#recipes
