# Docker Registry Module Configuration

## Overview

The Docker Registry module has its own operator (Docker Registry Operator). It watches the Docker Registry custom resource (CR) and reconfigures (reconciles) the Docker Registry workloads.

The DockerRegistry CR allows you to store images in five ways: filesystem, Azure, S3, GCS, and BTP Object Store, each requiring specific configurations. See [Registry Storage Configuration](00-30-storage-configuration.md) to learn more.

The DockerRegistry CR is the API to configure the Docker Registry module.

The default configuration of the Docker Registry module is the following:

```yaml
apiVersion: operator.kyma-project.io/v1alpha1
kind: DockerRegistry
metadata:
  name: default
  namespace: docker-registry
spec: {}
```

## Configure Resources

You can set CPU and memory limits and requests for the Docker Registry container using the `resources` field. If you don't provide any configuration, the operator uses the default values.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `resources.limits.cpu` | Maximum CPU allocated to the container | `400m` |
| `resources.limits.memory` | Maximum memory allocated to the container | `800Mi` |
| `resources.requests.cpu` | CPU reserved for the container | `10m` |
| `resources.requests.memory` | Memory reserved for the container | `300Mi` |

### Example

```yaml
apiVersion: operator.kyma-project.io/v1alpha1
kind: DockerRegistry
metadata:
  name: default
  namespace: docker-registry
spec:
  resources:
    limits:
      cpu: 500m
      memory: 1Gi
    requests:
      cpu: 100m
      memory: 256Mi
```

## Configure Replicas

You can increase the number of Docker Registry Pod replicas using the `replicas` field. The minimum and default value is `1`. Increase this value in high-concurrency environments where multiple clients push images simultaneously.

### Example

```yaml
apiVersion: operator.kyma-project.io/v1alpha1
kind: DockerRegistry
metadata:
  name: default
  namespace: docker-registry
spec:
  replicas: 3
```

## Configure Logging

You can configure logging for the Docker Registry Pods using the `logging` field.

| Parameter | Description | Valid values | Default |
|-----------|-------------|--------------|---------|
| `logging.level` | Log verbosity level | `error`, `warn`, `info`, `debug` | `info` |
| `logging.format` | Log output format | `json`, `text`, `console` | `json` |
| `logging.accessLogEnabled` | Enable HTTP access logs | `true`, `false` | `false` |

> [!NOTE]
> The `console` format is an alias for `text`.
> Access logs use Apache Combined Log Format and cannot use the configured formatter. Disable them for consistent log output.

### Example

```yaml
apiVersion: operator.kyma-project.io/v1alpha1
kind: DockerRegistry
metadata:
  name: default
  namespace: docker-registry
spec:
  logging:
    level: debug
    format: json
    accessLogEnabled: false
```

## Configure Docker Registry Operator Logging

To update the operator's logging configuration, edit the `dockerregistry-operator-config` ConfigMap in the `docker-registry` namespace.

### Change Log Level and Format

```bash
kubectl patch configmap dockerregistry-operator-config -n docker-registry --type merge -p '{"data":{"log-config.yaml":"logLevel: debug\nlogFormat: console"}}'
```

> [!NOTE]
> It is not possible to dynamically change the log format for the Docker Registry Operator. To change it, update the ConfigMap and restart the Pods.
