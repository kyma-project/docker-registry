# Docker Registry Network Policies

Learn about the network policies for the Docker Registry module.

## Overview

The Docker Registry module defines network policies to ensure communication within the Kubernetes cluster, particularly in environments where a deny-all network policy is applied.

When a cluster-wide deny-all network policy is enforced, which blocks all ingress and egress traffic by default, the Docker Registry network policies explicitly allow only the necessary communication paths to ensure the module functions correctly.

**Network Policies for the Docker Registry Module**

| Policy Name | Description |
|-------------|-------------|
| `kyma-project.io--dockerregistry-allow-registry-api` | Allows ingress to the Docker Registry API port (TCP 5000) from any source. This is the main interface through which clients push and pull container images. |
| `kyma-project.io--dockerregistry-allow-metrics-policy` | Allows ingress to the metrics endpoint (TCP 5001) from Pods labeled `app.kubernetes.io/instance: rma` or `networking.kyma-project.io/metrics-scraping: allowed` for metrics scraping. |
| `kyma-project.io--dockerregistry-allow-to-dns` | Allows egress to DNS services for cluster and external DNS resolution. Applies to any IP on port 53, and Pods labeled `k8s-app: kube-dns` or `k8s-app: node-local-dns` in the `kube-system` namespace on ports 53 and 8053. |
| `kyma-project.io--dockerregistry-allow-to-all` | Allows unrestricted outbound traffic from Docker Registry Pods to any destination. Applied only when an external storage backend is configured (Azure, S3, GCP, or BTP Object Store). Not applied when filesystem storage is used. |
| `kyma-project.io--dockerregistry-operator-allow-to-apiserver` | Allows egress from the Docker Registry Operator Pods to the Kubernetes API server (TCP 443, 6443). |
| `kyma-project.io--dockerregistry-operator-allow-to-dns` | Allows egress from the Docker Registry Operator Pods to DNS services for cluster and external DNS resolution. Targets any IP on port 53, and Pods labeled `k8s-app: kube-dns` or `k8s-app: node-local-dns` in the `kube-system` namespace on ports 53 and 8053. |

## Verify Status

To list the network policies belonging to the Docker Registry module, run the following command:

```bash
kubectl get networkpolicies -n kyma-system -l kyma-project.io/module=docker-registry
```
