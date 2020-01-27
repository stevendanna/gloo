#!/bin/sh
set -e

if [ -n "$ENVOY_SIDECAR" ]
then
  /usr/local/bin/envoy -c /etc/envoy/envoy-sidecar.yaml
else
  /usr/local/bin/envoyinit
fi
exec "$@"