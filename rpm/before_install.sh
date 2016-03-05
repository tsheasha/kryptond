#!/bin/bash

USER="relay"

id $USER > /dev/null 2>&1
if [ $? != 0 ]; then
  useradd --no-create-home --system --user-group $USER
fi

mkdir -p /var/log/relayd
chown relay:relay /var/log/relayd
