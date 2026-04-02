# ClusterRoles

Learn about ClusterRoles in the Docker Registry module. 

The Docker Registry module includes several ClusterRoles that are used to manage permissions for the Docker Registry Operator and to aggregate permissions for end users.

## Docker Registry Edit ClusterRole

With the `kyma-docker-registry-edit` ClusterRole, you can edit the Docker Registry resources. For the available options, see the following table:

| API Group | Resources | Verbs |
|---|---|---|
| `operator.kyma-project.io` | `dockerregistries` | `create`, `delete`, `get`, `list`, `patch`, `update`, `watch` |
| `operator.kyma-project.io` | `dockerregistries/status` | `get` |

## Docker Registry View ClusterRole

With the `kyma-docker-registry-view` ClusterRole, you can view the Docker Registry resources. For the available options, see the following table:

| API Group | Resources | Verbs |
|---|---|---|
| `operator.kyma-project.io` | `dockerregistries` | `get`, `list`, `watch` |
| `operator.kyma-project.io` | `dockerregistries/status` | `get` |

## Role Aggregation

The Docker Registry module uses the Kubernetes [role aggregation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#aggregated-clusterroles) to automatically extend the default `edit` and `view` ClusterRoles with Docker Registry-specific permissions.

- **kyma-docker-registry-edit**: Aggregated to `edit` ClusterRole
- **kyma-docker-registry-view**: Aggregated to `view` ClusterRole

This means that if you have the default Kubernetes `edit` or `view` ClusterRoles, you automatically receive the corresponding Docker Registry permissions without requiring additional role bindings.
