#!/bin/sh
set -e

echo "Starting up SDS Server..."
/usr/local/bin/sds &

echo "Starting Envoy..."
/usr/local/bin/envoy -c /etc/envoy.yaml