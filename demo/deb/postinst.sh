#!/bin/sh -xe

echo "post inst"

addgroup --system hello
adduser --system hello --no-create-home --home /nonexistent

if [ -f "/etc/init.d/hello.sh" ]; then
  update-rc.d hello defaults
  invoke-rc.d hello start
fi
if [ -f "/lib/systemd/system/hello.service" ]; then
  systemctl daemon-reload
  systemctl enable hello.service
  systemctl start hello.service
fi
