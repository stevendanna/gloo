#!/usr/bin/env bash

GLOO_VERSION='1.2.15'

# Will exit script if we would use an uninitialised variable (nounset) or when a
# simple command (not a control structure) fails (errexit)
set -eu

trap 'kill $(jobs -p)' EXIT

# Get directory this script is located in to access script local files
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"

if [[ "${OSTYPE}" == "darwin"* ]]; then
  # need to install and run Weave Net https://www.weave.works/docs/net/latest/install/installing-weave/
  if [[ -x $(command -v weave) ]]; then
    WEAVE_CMD="$(command -v weave)"
  else
    curl -L git.io/weave -o "${SCRIPT_DIR}/weave"
    chmod a+x "${SCRIPT_DIR}/weave"
    WEAVE_CMD="${SCRIPT_DIR}/weave"
  fi

  "${WEAVE_CMD}" stop
  "${WEAVE_CMD}" launch || true # ignore errors as 2nd+ call returns error saying already running
fi

echo "Launching Consul-Vault-Nomad for Demo (see ${SCRIPT_DIR}/.hashi-logs for logs)"
("${SCRIPT_DIR}/launch-consul-vault-nomad-dev.sh" >"${SCRIPT_DIR}/.hashi-logs" 2>&1) &

sleep 10

if ! [[ -x $(command -v levant) ]]; then
  printf '\nYou must install levant first\n\n'
  exit
fi

VARFILE="${SCRIPT_DIR}/variables.yaml"

INGRESS_IP='localhost'
DOCKER_HOST='localhost'
if [[ "${OSTYPE}" == "linux-gnu" ]]; then
  DOCKER_HOST='172.17.0.1'
  INGRESS_IP='172.17.0.1'
elif [[ "${OSTYPE}" == "darwin"* ]]; then
  DOCKER_HOST='host.docker.internal'
fi

cat >"${SCRIPT_DIR}/variables.yaml" <<EOF
datacenter: dc1

config:
  # the "namespace" where Gloo will read/write configuration
  # change this for multiple installations of Gloo
  namespace: gloo-system
  # the rate to poll Vault for secrets
  # maximum wait time on blocking requests to Consul
  refreshRate: 30s

consul:
  address: ${DOCKER_HOST}:8500

vault:
  address: http://${DOCKER_HOST}:8200
  token: root

gloo:
  # the port where Gloo serves config to Envoy
  xdsPort: 9977
  image:
    registry: quay.io/solo-io
    repository: gloo
    tag: ${GLOO_VERSION}
  cpuLimit: 1000
  memLimit: 500
  bandwidthLimit: 10
  # number of instances of gloo config server
  replicas: 1

discovery:
  image:
    registry: quay.io/solo-io
    repository: discovery
    tag: ${GLOO_VERSION}
  cpuLimit: 500
  memLimit: 500
  bandwidthLimit: 10

gateway:
  image:
    registry: quay.io/solo-io
    repository: gateway
    tag: ${GLOO_VERSION}
  cpuLimit: 250
  memLimit: 250
  bandwidthLimit: 5

gatewayProxy:
  image:
    registry: quay.io/solo-io
    repository: gloo-envoy-wrapper
    tag: ${GLOO_VERSION}
  cpuLimit: 500
  memLimit: 500
  bandwidthLimit: 100
  # number of instances of gateway proxy
  replicas: 1
  httpPort: 8080
  httpsPort: 8443
  adminPort: 19000
  # expose the http and https ports
  # on the host machine
  exposePorts: true
EOF

if [[ "${OSTYPE}" == "darwin"* ]]; then
  cat >>"${SCRIPT_DIR}/variables.yaml" <<EOF

# use this network plugin when running on mac
dockerNetwork: weave
EOF
fi

printf '\nDeploying Gloo\n\n'
levant deploy \
  -var-file="${VARFILE}" \
  jobs/gloo.nomad

printf '\nDeploying Petstore\n\n'
levant deploy \
  -var-file="${VARFILE}" \
  jobs/petstore.nomad

FAIL=0

printf '\nAdding route to Petstore\n\n'
glooctl add route \
  --path-prefix='/' \
  --dest-name='petstore' \
  --prefix-rewrite='/api/pets' \
  --use-consul

sleep 5

printf '\ncURL the gateway\n\n'
curl "${INGRESS_IP}:8080/"

printf '\nCtrl+C to exit.\n\n'

for job in $(jobs -p); do
  echo "${job}"
  wait "${job}" || ((FAIL++))
done

echo "${FAIL} background processes failed"
