# Docker Registry Module Configuration

## Overview

The Docker Registry module has its own operator (Docker Registry Operator). It watches the Docker Registry custom resource (CR) and reconfigures (reconciles) the Docker Registry workloads.

The DockerRegistry CR allows you to store images in five ways: filesystem, Azure, s3, GCP, and BTP Object Store, each requiring specific configurations. See [Reigstry Storage Configuration](00-30-storage-configuration.md) to learn more.

The Docker Registry CR becomes an API to configure the Docker Registry module.

The default configuration of the Docker Registry module is the following:

   ```yaml
   apiVersion: operator.kyma-project.io/v1alpha1
   kind: DockerRegistry
   metadata:
     name: default
     namespace: kyma-system
   spec: {}

   ```

## Docker Registry Logging

You can configure logging for the Docker Registry Pods using the `logging` section in the DockerRegistry CR spec.

### Configuration Options

| Parameter | Description | Valid Values | Default |
|-----------|-------------|--------------|---------|
| `level` | Log verbosity level | `error`, `warn`, `info`, `debug` | `info` |
| `format` | Log output format | `json`, `text`, `console` | `json` |
| `accessLogDisabled` | Disable HTTP access logs | `true`, `false` | `false` |

> [!NOTE]
> The `console` format is an alias for `text`.
> Access logs use Apache Combined Log Format and cannot use the configured formatter, so you may want to disable them for consistent log output.

### Example

```yaml
apiVersion: operator.kyma-project.io/v1alpha1
kind: DockerRegistry
metadata:
  name: default
  namespace: kyma-system
spec:
  logging:
    level: debug
    format: json
    accessLogDisabled: true
```

## Docker Registry Operator Logging Configuration

To update Operator's logging configuration, you can edit the `dockerregistry-operator-config` ConfigMap in the `kyma-system` namespace.

### Change log level and format
kubectl patch configmap dockerregistry-operator-config -n kyma-system --type merge -p '{"data":{"log-config.yaml":"logLevel: debug\nlogFormat: console"}}'

> [!NOTE]
> It is not possible to dynamically change the log format for the Docker Registry Operator. If you want to change it, update the ConfigMap and restart the Pods.