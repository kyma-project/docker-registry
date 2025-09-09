# Expose Docker Registry

This tutorial shows how you can expose the registry to the outside of the cluster with Istio.

## Prerequsities

* [kubectl](https://kubernetes.io/docs/tasks/tools/)
* [Kyma CLI](https://github.com/kyma-project/cli)
* [Docker](https://www.docker.com/)

## Steps

1. Expose the registry service by changing the **spec.externalAccess.enabled** flag to `true`:

    ```bash
    kubectl apply -n kyma-system -f - <<EOF
    apiVersion: operator.kyma-project.io/v1alpha1
    kind: DockerRegistry
    metadata:
      name: default
      namespace: kyma-system
    spec:
      externalAccess:
        enabled: true
    EOF
    ```

   Once the DockerRegistry CR becomes `Ready`, you will see a Secret name that you must use as `ImagePullSecret` when scheduling workloads in the cluster.

    ```yaml
    ...
    status:
      externalAccess:
        enabled: "True"
      internalAccess:
        enabled: "True"
        ...
        secretName: dockerregistry-config
    ```

2. Generate `config.json` file for docker-cli:

    ```bash
    kyma registry config-external --output config.json
    ```

3. Rename the image to contain the registry address:

    ```bash
    export REGISTRY_ADDRESS=$(kyma registry config-external --push-reg-addr)
    export IMAGE_NAME=<IMAGE_NAME> # put your image name here
    docker tag "${IMAGE_NAME}" "${REGISTRY_ADDRESS}/${IMAGE_NAME}"
    ```

4. Push the image to the registry:

    ```bash
    docker --config . push "${REGISTRY_ADDRESS}/${IMAGE_NAME}"
    ```

5. Create a Pod using the image from Docker Registry:

    ```bash

    export REGISTRY_INTERNAL_PULL_ADDRESS=$(kyma registry config-internal --pull-reg-addr)
    kubectl run my-pod --image="${REGISTRY_INTERNAL_PULL_ADDRESS}/${IMAGE_NAME}" --overrides='{ "spec": { "imagePullSecrets": [ { "name": "dockerregistry-config" } ] } }'
    ```
