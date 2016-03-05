#!/bin/bash

service 'relayd' stop

if [ "$(id 'relay')" ]; then
  userdel 'relay'
fi

rm -rf /var/log/relayd
