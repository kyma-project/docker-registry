#!/usr/bin/env bash

# standard bash error handling
set -o nounset  # treat unset variables as an error and exit immediately.
set -o errexit  # exit immediately when a command fails.
set -E          # needs to be set if we want the ERR trap
set -o pipefail # prevents errors in a pipeline from being masked

# Expected variables:
GITHUB_TOKEN=${GITHUB_TOKEN?"Define GITHUB_TOKEN env"} # github token used to upload the template yaml
RELEASE_ID=${RELEASE_ID?"Define RELEASE_ID env"} # github token used to upload the template yaml

uploadFile() {
  filePath=${1}
  ghAsset=${2}

  echo "Uploading ${filePath} as ${ghAsset}"
  response=$(curl -s -o output.txt -w "%{http_code}" \
                  --request POST --data-binary @"$filePath" \
                  -H "Authorization: token $GITHUB_TOKEN" \
                  -H "Content-Type: text/yaml" \
                   $ghAsset)
  if [[ "$response" != "201" ]]; then
    echo "Unable to upload the asset ($filePath): "
    echo "HTTP Status: $response"
    cat output.txt
    exit 1
  else
    echo "$filePath uploaded"
  fi
}

# dev registry is used because even if --dry-run is set, the cli expects the --registry flag to check connection
modulectl create -c module-config.yaml --registry https://europe-docker.pkg.dev/kyma-project/dev --dry-run -o module-template.yaml

echo "Generated module-template.yaml:"
cat module-template.yaml

make -C components/operator/ render-manifest

echo "Generated dockerregistry-operator.yaml:"
cat dockerregistry-operator.yaml

echo "Updating github release with assets"
UPLOAD_URL="https://uploads.github.com/repos/kyma-project/docker-registry/releases/${RELEASE_ID}/assets"

uploadFile "dockerregistry-operator.yaml" "${UPLOAD_URL}?name=dockerregistry-operator.yaml"
uploadFile "config/samples/default-dockerregistry-cr.yaml" "${UPLOAD_URL}?name=default-dockerregistry-cr.yaml"
uploadFile "module-template.yaml" "${UPLOAD_URL}?name=module-template.yaml"
