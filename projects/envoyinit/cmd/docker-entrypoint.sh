#!/bin/sh
set -e

if [ -n "$ENVOY_SIDECAR" ]
then
  echo "Starting Envoy..."
  /usr/local/bin/envoy -c /etc/envoy/envoy-sidecar.yaml &
  sleep 1 # Wait for Envoy to start up properly
  echo "Starting up SDS Server..."
  /usr/local/bin/sds
else
  /usr/local/bin/envoyinit
fi
exec "$@"