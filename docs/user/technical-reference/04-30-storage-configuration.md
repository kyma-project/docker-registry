# Registry Storage Configuration

The DockerRegistry CR allows you to store images in three ways: filesystem, Azure, and s3. This document describes how to configure DockerRegistry CR to cooperate with all these storage types.

## Filesystem

The filesystem storage is a built-in storage type based on the PersistentVolumeClaim CR, which is part of the Kubernetes functionality. This is a default DockerRegistry CR configuration, and no additional configuration is needed.

All images pushed to this storage are removed when the Docker Registry is uninstalled or the cluster is removed. Stored images can't be shared between clusters.

### Sample CR

```yaml
apiVersion: operator.kyma-project.io/v1alpha1
kind: DockerRegistry
metadata:
    name: default
    namespace: kyma-system
spec: {}
```

## Azure

The Azure storage can be configured in the DockerRegistry `spec.storage.azure` field. The only thing that is required is the `secretName` field that must contain the name of the Secret with Azure configuration located in the same namespace. The following Secret must have three values inside:

* `container` - contains the name of the storage container
* `accountKey` - contains the key used to authenticate to the Azure Storage
* `accountName` - contains the name used to authenticate to the Azure Storage

The images can be stored centrally and shared between clusters so that different registries can reuse specific layers or whole images. After deleting the cluster or uninstalling the registry module, images will not be removed.

### Sample CR

```yaml
apiVersion: operator.kyma-project.io/v1alpha1
kind: DockerRegistry
metadata:
    name: default
    namespace: kyma-system
spec:
    storage:
        azure:
            secretName: azure-storage
```

### Sample Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: azure-storage
  namespace: kyma-system
data:
  accountKey: "YWNjb3VudEtleQ=="
  accountName: "YWNjb3VudE5hbWU="
  container: "Y29udGFpbmVy"
```

## s3

Similarly to Azure, the s3 storage can be configured in the DockerRegistry `spec.storage.s3` field. The only required field is `bucket` which contains the s3 bucket name and `region` which specifies where the bucket is located. This storage type allows you to provide additional optional configuration, which is described in [DockerRegistry CR](../resources/06-20-docker-registry-cr.md). One of the optional configurations is the `secretName` that contains the authentication method to the s3 storage in the following format:

* `accountKey` - contains the key used to authenticate to the s3 storage
* `secretKey` - contains the name used to authenticate to the s3 storage

### Sample CR

```yaml
apiVersion: operator.kyma-project.io/v1alpha1
kind: DockerRegistry
metadata:
  name: default
  namespace: kyma-system
spec:
  storage:
    s3:
      bucket: "bucketName"
      region: "eu-central-1"
      regionEndpoint: "s3-eu-central-1.amazonaws.com"
      encrypt: false
      secure: true
      secretName: "s3-storage"
```

### Sample Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: s3-storage
  namespace: kyma-system
data:
  accessKey: "YWNjZXNzS2V5"
  secretKey: "c2VjcmV0S2V5"
```

## GCS

Google cloud storage can be configured using the `spec.storage.gcs` field. The only required field is `bucket` which contains the GCS bucket name. This storage type allows you to provide additional optional configuration, which is described in [DockerRegistry CR](../resources/06-20-docker-registry-cr.md). One of the optional configurations is the `secretName` that contains the authentication method to the GCS storage which is a private service account key in JSON format.

### Sample CR

```yaml
apiVersion: operator.kyma-project.io/v1alpha1
kind: DockerRegistry
metadata:
  name: default
  namespace: kyma-system
spec:
  storage:
    gcs:
      bucket: "bucketName"
      secretName: "gcs-secret"
      rootDirectory: "dir"
      chunkSize: 5242880
```

### Sample Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: gcs-secret
  namespace: kyma-system
data:
   key: "Z3Njc2VjcmV0"
```