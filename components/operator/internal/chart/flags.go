package chart

import (
	"fmt"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/strvals"
)

type FlagsBuilder interface {
	Build() (map[string]interface{}, error)
	WithFullname(fullname string) *flagsBuilder
	WithRegistryCredentials(username string, password string) *flagsBuilder
	WithRegistryHttpSecret(httpSecret string) *flagsBuilder
	WithServicePort(servicePort int64) *flagsBuilder
	WithVirtualService(host, gateway string) *flagsBuilder
	WithNodePort(nodePort int64) *flagsBuilder
	WithPVCDisabled() *flagsBuilder
	WithAzure(secret *v1alpha1.StorageAzureSecrets) *flagsBuilder
	WithS3(config *v1alpha1.StorageS3, secret *v1alpha1.StorageS3Secrets) *flagsBuilder
	WithFilesystem() *flagsBuilder
	WithDeleteEnabled(bool) *flagsBuilder
	WithPVC(config *v1alpha1.StoragePVC) *flagsBuilder
	WithGCS(config *v1alpha1.StorageGCS, secret *v1alpha1.StorageGCSSecrets) *flagsBuilder
	WithManagedByLabel(string) *flagsBuilder
}

type flagsBuilder struct {
	flags map[string]interface{}
}

func NewFlagsBuilder() FlagsBuilder {
	return &flagsBuilder{
		flags: map[string]interface{}{},
	}
}

func (fb *flagsBuilder) Build() (map[string]interface{}, error) {
	flags := map[string]interface{}{}
	for key, value := range fb.flags {
		flag := fmt.Sprintf("%s=%v", key, value)
		err := strvals.ParseInto(flag, flags)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse %s flag", flag)
		}
	}
	return flags, nil
}

func (fb *flagsBuilder) WithFullname(fullname string) *flagsBuilder {
	fb.flags["FullnameOverride"] = fullname
	return fb
}

func (fb *flagsBuilder) WithRegistryCredentials(username, password string) *flagsBuilder {
	fb.flags["dockerRegistry.username"] = username
	fb.flags["dockerRegistry.password"] = password
	return fb
}

func (fb *flagsBuilder) WithRegistryHttpSecret(httpSecret string) *flagsBuilder {
	fb.flags["registryHTTPSecret"] = httpSecret
	return fb
}

func (fb *flagsBuilder) WithServicePort(servicePort int64) *flagsBuilder {
	fb.flags["service.port"] = servicePort
	fb.flags["configData.http.addr"] = fmt.Sprintf(":%d", servicePort)
	return fb
}

func (fb *flagsBuilder) WithVirtualService(host, gateway string) *flagsBuilder {
	fb.flags["virtualService.enabled"] = true
	fb.flags["virtualService.host"] = host
	fb.flags["virtualService.gateway"] = gateway
	return fb
}

func (fb *flagsBuilder) WithNodePort(nodePort int64) *flagsBuilder {
	fb.flags["registryNodePort"] = nodePort
	return fb
}

func (fb *flagsBuilder) WithPVCDisabled() *flagsBuilder {
	fb.flags["persistence.enabled"] = false
	return fb
}

func (fb *flagsBuilder) WithAzure(secret *v1alpha1.StorageAzureSecrets) *flagsBuilder {
	fb.flags["storage"] = "azure"
	fb.flags["secrets.azure.accountName"] = secret.AccountName
	fb.flags["secrets.azure.accountKey"] = secret.AccountKey
	fb.flags["secrets.azure.container"] = secret.Container
	return fb
}

func (fb *flagsBuilder) WithS3(config *v1alpha1.StorageS3, secret *v1alpha1.StorageS3Secrets) *flagsBuilder {
	fb.flags["storage"] = "s3"

	fb.flags["s3.bucket"] = config.Bucket
	fb.flags["s3.region"] = config.Region
	fb.flags["s3.encrypt"] = config.Encrypt
	fb.flags["s3.secure"] = config.Secure

	if config.RegionEndpoint != "" {
		fb.flags["s3.regionEndpoint"] = config.RegionEndpoint
	}

	if secret != nil {
		fb.flags["secrets.s3.accessKey"] = secret.AccessKey
		fb.flags["secrets.s3.secretKey"] = secret.SecretKey
	}

	return fb
}

func (fb *flagsBuilder) WithDeleteEnabled(enabled bool) *flagsBuilder {
	fb.flags["configData.storage.delete.enabled"] = enabled
	// restart deployment registry deploy to fetch new configuration from configmap
	// set rollme value to the contant value (reason) to not restart deployment on every reconciliation
	return fb.withRollme(fmt.Sprintf("configData.storage.delete.enabled=%t", enabled))
}

func (fb *flagsBuilder) WithFilesystem() *flagsBuilder {
	fb.flags["storage"] = "filesystem"
	fb.flags["configData.storage.filesystem.rootdirectory"] = "/var/lib/registry"
	return fb
}

func (fb *flagsBuilder) WithPVC(config *v1alpha1.StoragePVC) *flagsBuilder {
	fb.flags["persistence.enabled"] = true
	fb.flags["persistence.existingClaim"] = config.Name
	return fb
}

func (fb *flagsBuilder) WithGCS(config *v1alpha1.StorageGCS, secret *v1alpha1.StorageGCSSecrets) *flagsBuilder {
	fb.flags["storage"] = "gcs"
	fb.flags["gcs.bucket"] = config.Bucket

	if config.Rootdirectory != "" {
		fb.flags["gcs.rootdirectory"] = config.Rootdirectory
	}

	if config.Chunksize != 0 {
		fb.flags["gcs.chunkSize"] = config.Chunksize
	}

	if secret != nil {
		fb.flags["secrets.gcs.accountkey"] = secret.AccountKey
	}

	return fb
}

func (fb *flagsBuilder) WithManagedByLabel(managedBy string) *flagsBuilder {
	fb.flags["commonLabels.app\\.kubernetes\\.io/managed-by"] = managedBy
	return fb
}

// withRollme allows to set custom values for the `rollme` field in chart
// it merges values for many command executions in format <value1>,<value2>,...,<valueN>
func (fb *flagsBuilder) withRollme(value string) *flagsBuilder {
	rollme, ok := fb.flags["rollme"]
	if ok {
		value = fmt.Sprintf("%v\\,%s", rollme, value)
	}

	fb.flags["rollme"] = value
	return fb
}
