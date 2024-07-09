# Docker Registry

## Status

![GitHub tag checks state](https://img.shields.io/github/checks-status/kyma-project/docker-registry/main?label=docker-registry&link=https%3A%2F%2Fgithub.com%2Fkyma-project%2Fdocker-registry%2Fcommits%2Fmain)
[![REUSE status](https://api.reuse.software/badge/github.com/kyma-project/docker-registry)](https://api.reuse.software/info/github.com/kyma-project/docker-registry)

## Overview

Docker Registry module aims to provide a lightweight, open-source Docker registry for storing and distributing container images in kubernetes environment. It could be used in self-contained development clusters or testing environments allowing you to test deployments in an isolated system. 
Usage on production clusters, where a full-fledged, highly-available, production-grade registry is necessary, is not recomended.

Docker Registry module comes with a dedicated kubernetes operator for efortless installation and upgrades. It manages the lifecycle of the Docker Registry resources according to the desired specification expressed via dedicated `Dockerregistry` Custom Resource. Read more about configuration options in the [user documentation](./docs/user). 

## Install

1. Create the `kyma-system` namespace:

```bash
kubectl create namespace kyma-system
```

2. Apply the following script to install Docker Registry Operator:

```bash
kubectl apply -f https://github.com/kyma-project/docker-registry/releases/latest/download/dockerregistry-operator.yaml
```

3. To get Docker Registry installed, apply the sample Docker Registry custom resource (CR):

```bash
kubectl apply -f https://github.com/kyma-project/docker-registry/releases/latest/download/default-dockerregistry-cr.yaml
```

## Usage

Having installed Docker Registry module, you can start pushing images to it via [kyma CLI (v3)](https://github.com/kyma-project/cli?tab=readme-ov-file#install) or, if you decide to expose the registry externally, using Docker CLI.
Please find both variants described in the dedicated tutorials:
 - [How to use non exposed registry](docs/user/tutorials/01-10-simple-usage.md)
 - [How to Expose Registry](docs/user/tutorials/01-20-expose-registry.md)

## Development

### Prerequisites

- Access to a Kubernetes (v1.24 or higher) cluster
- [Go](https://go.dev/)
- [k3d](https://k3d.io/)
- [Docker](https://www.docker.com/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Kubebuilder](https://book.kubebuilder.io/)

## Installation in the k3d Cluster Using Make Targets

1. Clone the project.

    ```bash
    git clone https://github.com/kyma-project/docker-registry.git && cd docker-registry/
    ```

2. Build Docker Registry Operator locally and run it in the k3d cluster.

    ```bash
    make run
    ```

> **NOTE:** To clean up the k3d cluster, use the `make delete-k3d` make target.

## Using Docker Registry Operator

- Create a Docker Registry instance.

    ```bash
    kubectl apply -f config/samples/default-dockerregistry-cr.yaml
    ```

- Delete a Docker Registry instance.

    ```bash
    kubectl delete -f config/samples/default-dockerregistry-cr.yaml
    ```
