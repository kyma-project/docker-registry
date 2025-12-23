package flags

import (
	"fmt"
	"strings"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/manager-toolkit/installation/chart"
)

const (
	FullnameOverride = "dockerregistry"
)

type Builder struct {
	chart.FlagsBuilder
	rollmeValues []string
}

func NewBuilder() *Builder {
	return &Builder{
		FlagsBuilder: chart.NewFlagsBuilder(),
	}
}

func (fb *Builder) WithFullname(fullname string) *Builder {
	fb.With("FullnameOverride", fullname)
	return fb
}

func (fb *Builder) WithRegistryCredentials(username, password string) *Builder {
	fb.With("dockerRegistry.username", username)
	fb.With("dockerRegistry.password", password)
	return fb
}

func (fb *Builder) WithRegistryHttpSecret(httpSecret string) *Builder {
	fb.With("registryHTTPSecret", httpSecret)
	return fb
}

func (fb *Builder) WithServicePort(servicePort int64) *Builder {
	fb.With("service.port", servicePort)
	fb.With("configData.http.addr", fmt.Sprintf(":%d", servicePort))
	return fb
}

func (fb *Builder) WithVirtualService(host, gateway string) *Builder {
	fb.With("virtualService.enabled", true)
	fb.With("virtualService.host", host)
	fb.With("virtualService.gateway", gateway)
	return fb
}

func (fb *Builder) WithNodePort(nodePort int64) *Builder {
	fb.With("registryNodePort", nodePort)
	return fb
}

func (fb *Builder) WithPVCDisabled() *Builder {
	fb.With("persistence.enabled", false)
	return fb
}

func (fb *Builder) WithAzure(secret *v1alpha1.StorageAzureSecrets) *Builder {
	fb.With("storage", "azure")
	fb.With("secrets.azure.accountName", secret.AccountName)
	fb.With("secrets.azure.accountKey", secret.AccountKey)
	fb.With("secrets.azure.container", secret.Container)
	return fb
}

func (fb *Builder) WithS3(config *v1alpha1.StorageS3, secret *v1alpha1.StorageS3Secrets) *Builder {
	fb.With("storage", "s3")
	fb.With("s3.bucket", config.Bucket)
	fb.With("s3.region", config.Region)
	fb.With("s3.encrypt", config.Encrypt)
	fb.With("s3.secure", config.Secure)

	if config.RegionEndpoint != "" {
		fb.With("s3.regionEndpoint", config.RegionEndpoint)
	}

	if secret != nil {
		fb.With("secrets.s3.accessKey", secret.AccessKey)
		fb.With("secrets.s3.secretKey", secret.SecretKey)
	}

	return fb
}

func (fb *Builder) WithDeleteEnabled(enabled bool) *Builder {
	fb.With("configData.storage.delete.enabled", enabled)
	// restart deployment registry deploy to fetch new configuration from configmap
	// set rollme value to the constant value (reason) to not restart deployment on every reconciliation
	return fb.withRollme(fmt.Sprintf("configData.storage.delete.enabled=%t", enabled))
}

func (fb *Builder) WithFilesystem() *Builder {
	fb.With("storage", "filesystem")
	fb.With("configData.storage.filesystem.rootdirectory", "/var/lib/registry")
	return fb
}

func (fb *Builder) WithPVC(config *v1alpha1.StoragePVC) *Builder {
	fb.With("persistence.enabled", true)
	fb.With("persistence.existingClaim", config.Name)
	return fb
}

func (fb *Builder) WithGCS(config *v1alpha1.StorageGCS, secret *v1alpha1.StorageGCSSecrets) *Builder {
	fb.With("storage", "gcs")
	fb.With("gcs.bucket", config.Bucket)

	if config.Rootdirectory != "" {
		fb.With("gcs.rootdirectory", config.Rootdirectory)
	}

	if config.Chunksize != 0 {
		fb.With("gcs.chunkSize", config.Chunksize)
	}

	if secret != nil {
		fb.With("secrets.gcs.accountkey", secret.AccountKey)
	}

	return fb
}

func (fb *Builder) WithManagedByLabel(managedBy string) *Builder {
	fb.With("commonLabels.app\\.kubernetes\\.io/managed-by", managedBy)
	return fb
}

// withRollme allows to set custom values for the `rollme` field in chart
// it merges values for many command executions in format <value1>,<value2>,...,<valueN>
func (fb *Builder) withRollme(value string) *Builder {
	fb.rollmeValues = append(fb.rollmeValues, value)
	fb.With("rollme", strings.Join(fb.rollmeValues, ","))
	return fb
}
