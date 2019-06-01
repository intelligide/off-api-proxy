#!/bin/sh

set -eu

exec su-exec "${PUID}:${PGID}" /bin/syncthing "$@"
