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

vagrant up cli

vagrant rsync cli && vagrant ssh cli -c "export GH_TOKEN=$GH_TOKEN; sh /vagrant/vagrant-run.sh"

vagrant destroy cli -f
```

# See also

https://github.com/mh-cbon/go-bin-deb#recipes
