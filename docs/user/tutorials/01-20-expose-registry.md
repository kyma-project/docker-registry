# Expose Registry

This tutorial shows how you can expose registry to the outside of the cluster with istio.

## Prerequsities

* [kubectl](https://kubernetes.io/docs/tasks/tools/)
* [Docker](https://www.docker.com/)

## Steps

1. Export cluster address:

    ```bash
    export CLUSTER_ADDRESS=<CLUSTER_ADDRESS>
    ```

    >**NOTE:** Put your cluster address instead of the `<CLUSTER_ADDRESS>`. You can find it in the `kyma-system/kyma-gateway` gateway resource.

1. Expose the registry service using a VirtualService CR based on the `kyma-gateway` gateway in the `kyma-system` namespace:

    ```bash
    kubectl apply -n kyma-system -f - <<EOF
    apiVersion: networking.istio.io/v1beta1
    kind: VirtualService
    metadata:
        name: registry-default-kyma-system
    spec:
        gateways:
        - kyma-system/kyma-gateway
        hosts:
        - registry-default-kyma-system.${CLUSTER_ADDRESS}
        http:
        - route:
            - destination:
                host: dockerregistry.kyma-system.svc.cluster.local
                port:
                    number: 5000
    EOF
    ```

2. Login to the registry using the docker-cli:

    ```bash
    export REGISTRY_USERNAME=$(kubectl get secrets -n kyma-system dockerregistry-config -o jsonpath={.data.username} | base64 -d)
    export REGISTRY_PASSWORD=$(kubectl get secrets -n kyma-system dockerregistry-config -o jsonpath={.data.password} | base64 -d)
    docker login -u ${REGISTRY_USERNAME} -p ${REGISTRY_PASSWORD} registry-default-kyma-system.${CLUSTER_ADDRESS}
    ```

3. Rename image to contains registry address:

    ```bash
    export IMAGE_NAME=<IMAGE_NAME> # put your image name here
    docker tag ${IMAGE_NAME} registry-default-kyma-system.${CLUSTER_ADDRESS}/${IMAGE_NAME}
    ```

4. Create registry auth secret:

    ```bash
    export REGISTRY_AUTH=$(echo -n "${REGISTRY_USERNAME}:${REGISTRY_PASSWORD}" | base64)
    export DOCKER_CONFIG_JSON=$(echo -n '{"auths": {"registry-default-kyma-system.'${CLUSTER_ADDRESS}'": {"auth": "'${REGISTRY_AUTH}'"}}}' | base64)

    kubectl apply -f - <<EOF
    apiVersion: v1
    kind: Secret
    metadata:
        name: exposed-registry-auth
    data:
        .dockerconfigjson: ${DOCKER_CONFIG_JSON}
    type: kubernetes.io/dockerconfigjson
    EOF
    ```

5. Push image to the registry:

    ```bash
    docker push registry-default-kyma-system.${CLUSTER_ADDRESS}/${IMAGE_NAME}
    ```

6. Create a Pod using the image from Docker Registry:

    ```bash
    kubectl run my-pod --image=registry-default-kyma-system.${CLUSTER_ADDRESS}/${IMAGE_NAME} --overrides='{ "spec": { "imagePullSecrets": [ { "name": "exposed-registry-auth" } ] } }'
    ```
