# Cluster Roles

The Docker Registry module includes several ClusterRoles that are used to manage permissions for the Docker Registry operator and to aggregate permissions for end users. This document describes all ClusterRoles bundled with the Docker Registry module.

## Docker Registry Edit ClusterRole

The `kyma-docker-registry-edit` ClusterRole allows users to edit Docker Registry resources.

| API Group | Resources | Verbs |
|---|---|---|
| `operator.kyma-project.io` | `dockerregistries` | `create`, `delete`, `get`, `list`, `patch`, `update`, `watch` |
| `operator.kyma-project.io` | `dockerregistries/status` | `get` |

## Docker Registry View ClusterRole

The `kyma-docker-registry-view` ClusterRole allows users to view Docker Registry resources.

| API Group | Resources | Verbs |
|---|---|---|
| `operator.kyma-project.io` | `dockerregistries` | `get`, `list`, `watch` |
| `operator.kyma-project.io` | `dockerregistries/status` | `get` |

## Role Aggregation

The Docker Registry module uses Kubernetes [role aggregation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#aggregated-clusterroles) to automatically extend the default `edit` and `view` ClusterRoles with Docker Registry-specific permissions.

- **kyma-docker-registry-edit**: Aggregated to `edit` ClusterRole
- **kyma-docker-registry-view**: Aggregated to `view` ClusterRole

This means that users who are granted the default Kubernetes `edit` or `view` ClusterRoles automatically receive the corresponding Docker Registry permissions without requiring additional role bindings.
