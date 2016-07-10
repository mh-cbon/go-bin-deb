#!/bin/sh -xe

echo "post rm"

if [ -f "/etc/init.d/hello.sh" ]; then
  invoke-rc.d hello stop
  update-rc.d hello disable
fi
if [ -f "/lib/systemd/system/hello.service" ]; then
  systemctl stop hello.service
  systemctl disable hello.service
fi
