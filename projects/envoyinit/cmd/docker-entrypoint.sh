#!/bin/sh
set -e

if [ -n "$ENVOY_SIDECAR" ]
then
  echo "Starting up SDS Server..."
  /usr/local/bin/sds &
  sleep 1 # Wait for SDS server to start up properly
  echo "Starting Envoy..."
  /usr/local/bin/envoy -c /etc/envoy/envoy-sidecar.yaml
else
  /usr/local/bin/envoyinit
fi
exec "$@"