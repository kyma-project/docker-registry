# Docker Registry Network Policies

## Overview

The Docker Registry module defines network policies to ensure communication within the Kubernetes cluster, particularly in environments where a deny-all network policy is applied. 
When a cluster-wide deny-all network policy is enforced, which blocks all ingress and egress traffic by default, the Docker Registry network policies explicitly allow only the necessary communication paths to ensure the module functions correctly.

## Network Policies

### Allow Registry API Access

**Name:** `kyma-project.io--dockerregistry-allow-registry-api`

**Purpose:** Enables external clients to interact with the Docker Registry API

**Details:**
- **Traffic Type:** Ingress
- **Port:** 5000 (TCP)
- **Description:** Allows inbound connections to the Docker Registry API port. This is the main interface through which clients push and pull container images. Without this policy, all external requests to the registry would be blocked.

**Use Case:** Essential for basic Docker Registry functionality, allowing Docker clients and Kubernetes operations to communicate with the registry service.


### Allow Metrics Scraping

**Name:** `kyma-project.io--dockerregistry-allow-metrics-policy`

**Purpose:** Allows the metrics collection agent to scrape metrics from the Docker Registry

**Details:**
- **Traffic Type:** Ingress
- **Port:** 5001 (TCP)
- **Allowed Sources:**
  - Pods with label `app.kubernetes.io/instance: rma` (Registry Metrics Agent)
  - Pods with label `networking.kyma-project.io/metrics-scraping: allowed`
- **Description:** Restricts access to the metrics endpoint (port 5001) only to authorized metrics collection agents. This ensures monitoring and observability of Docker Registry while maintaining security by preventing unauthorized access to operational metrics.

**Use Case:** Allows the Kyma metrics collection infrastructure to monitor Docker Registry performance and health without requiring network-wide ingress permissions.


### Allow DNS Queries

**Name:** `kyma-project.io--dockerregistry-allow-to-dns`

**Purpose:** Safeguards communication with Kubernetes DNS services

**Details:**
- **Traffic Type:** Egress
- **Ports:** 
  - Port 53 (TCP/UDP) - Standard DNS
  - Port 8053 (TCP/UDP) - Alternative DNS port
- **Allowed Destinations:**
  - Any IP address (0.0.0.0/0) on ports 53 TCP/UDP
  - Pods labeled `k8s-app: kube-dns` in `gardener.cloud/purpose: kube-system` namespace on ports 53 and 8053 (TCP/UDP)
  - Pods labeled `k8s-app: node-local-dns` in `gardener.cloud/purpose: kube-system` namespace on ports 53 and 8053 (TCP/UDP)
- **Description:** Enables Docker Registry to resolve internal and external domain names through Kubernetes DNS services. This allows the registry to discover storage backend endpoints, API services, and other internal Kubernetes services by hostname.

**Use Case:** Critical for service discovery within the cluster. Allows the registry to resolve DNS names for:
- Kubernetes API server
- Storage backend endpoints (when using external storage)
- Other internal services in the cluster


### Allow External Storage Backend Communication

**Name:** `kyma-project.io--dockerregistry-allow-to-all`

**Purpose:** Safeguards egress communication with external data backends when external storage is enabled

**Details:**
- **Traffic Type:** Egress
- **Scope:** All ports and destinations (unrestricted egress)
- **Conditional:** Only applied when the storage backend is **NOT** filesystem, for example, when using Azure, S3, GCP, or BTP Object Store
- **Description:** When external storage is configured, Docker Registry requires unrestricted outbound access to communicate with the external storage service. External storage backends (cloud object stores) use various ports and protocols, making it impractical to define granular rules. This policy permits all outbound traffic from Docker Registry Pods to any destination.

**Use Case:** Necessary when Docker Registry uses external storage options:
- **Amazon S3** - Requires HTTP/HTTPS connections to AWS S3 endpoints
- **Azure Blob Storage** - Requires HTTPS connections to Azure storage endpoints
- **Google Cloud Storage (GCP)** - Requires HTTPS connections to GCS endpoints
- **SAP BTP Object Store** - Requires HTTPS connections to BTP service endpoints
- **Filesystem storage** - Does not apply this policy (local storage only)
