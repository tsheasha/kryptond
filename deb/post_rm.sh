#!/bin/bash

set -e

case "$1" in
    purge)
        if [ "$(id 'relay')" ]; then
            userdel 'relay'
        fi
        rm -rf /var/log/relayd
        ;;
    remove|upgrade|failed-upgrade|abort-install|abort-upgrade|disappear)
            ;;
    *)
        echo "postrm called with unknown argument \`$1'" >&2
        exit 1
        ;;
esac

exit 0
