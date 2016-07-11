# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|

  config.vm.define "cli" do |cli|
    cli.vm.box = "debian/jessie64"
    cli.vm.synced_folder ".", "/vagrant"
    cli.vm.provider :virtualbox do |vb|
      vb.gui = true
    end
  end

  config.vm.define "gui" do |cli|
    cli.vm.box = "mrlesmithjr/jessie64-desktop-gnome"
    cli.vm.synced_folder ".", "/vagrant"
    cli.vm.provider :virtualbox do |vb|
      vb.customize ["modifyvm", :id, "--usb", "on"]
      vb.customize ["modifyvm", :id, "--usbehci", "off"]
      vb.gui = true
    end
  end

end
