
export GHTOKEN=`gh-api-cli get-auth -n release`

vagrant rsync cli && vagrant ssh cli -c "export GHTOKEN=$GHTOKEN; sh /vagrant/vagrant-run.sh"
