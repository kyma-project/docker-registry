# Docker Registry Network Policies

## Overview

The Docker Registry module defines network policies to ensure communication within the Kubernetes cluster, particularly in environments where a deny-all network policy is applied.
When a cluster-wide deny-all network policy is enforced, which blocks all ingress and egress traffic by default, the Docker Registry network policies explicitly allow only the necessary communication paths to ensure the module functions correctly.

## Network Policies

To list the network policies belonging to the Docker Registry module, run the following command:

```bash
kubectl get networkpolicies -n kyma-system -l kyma-project.io/module=docker-registry
```

The following tables describe the network policies for the Docker Registry module.

**Docker Registry Policies**

| Policy Name | Type | Port(s) | Description |
|-------------|------|---------|-------------|
| `kyma-project.io--dockerregistry-allow-registry-api` | Ingress | 5000 (TCP) | Allows inbound connections to the Docker Registry API port from any source. This is the main interface through which clients push and pull container images. |
| `kyma-project.io--dockerregistry-allow-metrics-policy` | Ingress | 5001 (TCP) | Allows ingress to the metrics endpoint from Pods labeled `app.kubernetes.io/instance: rma` or `networking.kyma-project.io/metrics-scraping: allowed` for metrics scraping. |
| `kyma-project.io--dockerregistry-allow-to-dns` | Egress | 53 (TCP/UDP), 8053 (TCP/UDP) | Allows egress to DNS services for cluster and external DNS resolution. Targets any IP on port 53, and Pods labeled `k8s-app: kube-dns` or `k8s-app: node-local-dns` in the `kube-system` namespace on ports 53 and 8053. |
| `kyma-project.io--dockerregistry-allow-to-all` | Egress | All | Allows unrestricted outbound traffic from Docker Registry Pods to any destination. Applied only when an external storage backend is configured (Azure, S3, GCP, or BTP Object Store). Not applied when filesystem storage is used. |

**Docker Registry Operator Policies**

| Policy Name | Type | Port(s) | Description |
|-------------|------|---------|-------------|
| `kyma-project.io--dockerregistry-operator-allow-to-apiserver` | Egress | 443 (TCP), 6443 (TCP) | Allows egress from the Docker Registry Operator Pods to the Kubernetes API server. |
| `kyma-project.io--dockerregistry-operator-allow-to-dns` | Egress | 53 (TCP/UDP), 8053 (TCP/UDP) | Allows egress from the Docker Registry Operator Pods to DNS services for cluster and external DNS resolution. Targets any IP on port 53, and Pods labeled `k8s-app: kube-dns` or `k8s-app: node-local-dns` in the `kube-system` namespace on ports 53 and 8053. |


