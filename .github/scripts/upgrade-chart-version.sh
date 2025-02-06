#!/usr/bin/env bash

# standard bash error handling
set -o nounset  # treat unset variables as an error and exit immediately.
set -o errexit  # exit immediately when a command fails.
set -E          # needs to be set if we want the ERR trap
set -o pipefail # prevents errors in a pipeline from being masked

# Expected variables:
MODULE_VERSION=${MODULE_VERSION?"define MODULE_VERSION env"} # module version used to set common labels

yq --inplace ".commonLabels.version=\"${MODULE_VERSION}\" | .commonLabels.\"app.kubernetes.io/version\"=\"${MODULE_VERSION}\"" ./config/docker-registry/values.yaml
