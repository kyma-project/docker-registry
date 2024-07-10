# Expose Registry

This tutorial shows how you can expose the registry to the outside of the cluster with Istio.

## Prerequsities

* [kubectl](https://kubernetes.io/docs/tasks/tools/)
* [Docker](https://www.docker.com/)

## Steps

1. Export cluster address:

    ```bash
    export CLUSTER_ADDRESS={YOUR_CLUSTER_ADDRESS}
    ```

    >[!NOTE] 
    > You can find your cluster address in the `kyma-system/kyma-gateway` gateway resource.

1. Expose the registry service by changing the **spec.externalAccess.enabled** flag to `true`. Optionally, you can also change the host name:

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
        hostPrefix: my-registry
    EOF
    ```
   
   Once the DockerRegistry CR becomes `Ready`, you see a Secret name that is used as `ImagePullSecret` when scheduling workloads in the cluster.
    ```yaml
    ...
    status:
      externalAccess:
        enabled: "True"
        ...
        secretName: dockerregistry-config-external
    ```

2. Log in to the registry using the docker-cli:

    ```bash
    export REGISTRY_USERNAME=$(kubectl get secrets -n kyma-system dockerregistry-config-external -o jsonpath={.data.username} | base64 -d)
    export REGISTRY_PASSWORD=$(kubectl get secrets -n kyma-system dockerregistry-config-external -o jsonpath={.data.password} | base64 -d)
    docker login -u ${REGISTRY_USERNAME} -p ${REGISTRY_PASSWORD} my-registry.${CLUSTER_ADDRESS}
    ```

3. Rename the image to contain the registry address:

    ```bash
    export IMAGE_NAME=<IMAGE_NAME> # put your image name here
    docker tag ${IMAGE_NAME} my-registry.${CLUSTER_ADDRESS}/${IMAGE_NAME}
    ```

4. Push the image to the registry:

    ```bash
    docker push my-registry.${CLUSTER_ADDRESS}/${IMAGE_NAME}
    ```

6. Create a Pod using the image from Docker Registry:

    ```bash
    kubectl run my-pod --image=my-registry.${CLUSTER_ADDRESS}/${IMAGE_NAME} --overrides='{ "spec": { "imagePullSecrets": [ { "name": "dockerregistry-config-external" } ] } }'
    ```
