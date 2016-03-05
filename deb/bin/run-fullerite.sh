#!/usr/bin/env bash

BINDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -L)"
RELAYD_DIR="$(dirname "${BINDIR}")"
RELAYD="${BINDIR}/relayd"
RELAYD_CONFIG="${RELAYD_DIR}/examples/config/relayd.conf.example"

ARGS="$@"
if [ -z "${ARGS}" ]; then
    ARGS="-c ${EXAMPLE_CONFIG}"
fi

exec ${RELAYD} ${ARGS}
