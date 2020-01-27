#!/bin/sh
set -e

if [ -n "$ENVOY_SIDECAR" ]
then
  echo "Starting up SDS Server..."
  /usr/local/bin/sds &
  echo "Starting Envoy..."
  /usr/local/bin/envoy -c /etc/envoy/envoy-sidecar.yaml
else
  /usr/local/bin/envoyinit
fi
exec "$@"