#!/bin/bash

if ! [ "$(id 'relay')" ]; then
  getent group 'relay' > /dev/null 2>&1
  exit_code=$?
  if [ $exit_code -eq 0 ]; then
    echo "creating user relay and adding to relay group"
    useradd --no-create-home --system -g"relay" "relay"
  elif [ $exit_code -eq 2 ]; then
    echo "creating user and group relay"
    useradd --no-create-home --system --user-group "relay"
  else
    echo "could not get group info, failed"
    exit 1
  fi
fi

echo "creating log directory: /var/log/relayd"
mkdir -p /var/log/relayd
chown relay:relay /var/log/relayd
