# Docker Registry Module

## What is the Docker Registry Module?

> [!WARNING]  
> Do not use Docker Registry in production clusters, where a full-fledged, highly-available, production-grade registry is necessary.

The Docker Registry module is a Kubernetes operator that adds the Docker Registry capability to a Kubernetes cluster. It installs packaged distribution images and configures them to be easily used in Kyma runtime.

## What is Docker Registry in Kyma?

Docker Registry brings extra capability to Kyma runtime, which is useful in the development phase. It allows developers to push their images into Kyma’s internal registry and use the images to spin workloads in the same cluster.

## Authorization

To assign access permissions to the Docker Registry module resources, use the following [aggregated ClusterRoles](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#aggregated-clusterroles):

- `kyma-docker-registry-view`
- `kyma-docker-registry-edit`

## Useful Links

If you want to perform some simple and more advanced tasks, check the [Docker Registry tutorials](tutorials/README.md).
