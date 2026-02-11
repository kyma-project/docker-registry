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
	_ = fb.With("FullnameOverride", fullname)
	return fb
}

func (fb *Builder) WithRegistryCredentials(username, password string) *Builder {
	_ = fb.With("dockerRegistry.username", username)
	_ = fb.With("dockerRegistry.password", password)
	return fb
}

func (fb *Builder) WithRegistryHttpSecret(httpSecret string) *Builder {
	_ = fb.With("registryHTTPSecret", httpSecret)
	return fb
}

func (fb *Builder) WithServicePort(servicePort int64) *Builder {
	_ = fb.With("service.port", servicePort)
	_ = fb.With("configData.http.addr", fmt.Sprintf(":%d", servicePort))
	return fb
}

func (fb *Builder) WithVirtualService(host, gateway string) *Builder {
	_ = fb.With("virtualService.enabled", true)
	_ = fb.With("virtualService.host", host)
	_ = fb.With("virtualService.gateway", gateway)
	return fb
}

func (fb *Builder) WithNodePort(nodePort int64) *Builder {
	_ = fb.With("registryNodePort", nodePort)
	return fb
}

func (fb *Builder) WithPVCDisabled() *Builder {
	_ = fb.With("persistence.enabled", false)
	return fb
}

func (fb *Builder) WithAzure(secret *v1alpha1.StorageAzureSecrets) *Builder {
	_ = fb.With("storage", "azure")
	_ = fb.With("secrets.azure.accountName", secret.AccountName)
	_ = fb.With("secrets.azure.accountKey", secret.AccountKey)
	_ = fb.With("secrets.azure.container", secret.Container)
	return fb
}

func (fb *Builder) WithS3(config *v1alpha1.StorageS3, secret *v1alpha1.StorageS3Secrets) *Builder {
	_ = fb.With("storage", "s3")
	_ = fb.With("s3.bucket", config.Bucket)
	_ = fb.With("s3.region", config.Region)
	_ = fb.With("s3.encrypt", config.Encrypt)
	_ = fb.With("s3.secure", config.Secure)

	if config.RegionEndpoint != "" {
		_ = fb.With("s3.regionEndpoint", config.RegionEndpoint)
	}

	if secret != nil {
		_ = fb.With("secrets.s3.accessKey", secret.AccessKey)
		_ = fb.With("secrets.s3.secretKey", secret.SecretKey)
	}

	return fb
}

func (fb *Builder) WithDeleteEnabled(enabled bool) *Builder {
	_ = fb.With("configData.storage.delete.enabled", enabled)
	// restart deployment registry deploy to fetch new configuration from configmap
	// set rollme value to the constant value (reason) to not restart deployment on every reconciliation
	return fb.withRollme(fmt.Sprintf("configData.storage.delete.enabled=%t", enabled))
}

func (fb *Builder) WithFilesystem() *Builder {
	_ = fb.With("storage", "filesystem")
	_ = fb.With("configData.storage.filesystem.rootdirectory", "/var/lib/registry")
	return fb
}

func (fb *Builder) WithPVC(config *v1alpha1.StoragePVC) *Builder {
	_ = fb.With("persistence.enabled", true)
	_ = fb.With("persistence.existingClaim", config.Name)
	return fb
}

func (fb *Builder) WithGCS(config *v1alpha1.StorageGCS, secret *v1alpha1.StorageGCSSecrets) *Builder {
	_ = fb.With("storage", "gcs")
	_ = fb.With("gcs.bucket", config.Bucket)

	if config.Rootdirectory != "" {
		_ = fb.With("gcs.rootdirectory", config.Rootdirectory)
	}

	if config.Chunksize != 0 {
		_ = fb.With("gcs.chunkSize", config.Chunksize)
	}

	if secret != nil {
		_ = fb.With("secrets.gcs.accountkey", secret.AccountKey)
	}

	return fb
}

func (fb *Builder) WithManagedByLabel(managedBy string) *Builder {
	_ = fb.With("commonLabels.app\\.kubernetes\\.io/managed-by", managedBy)
	return fb
}

func (fb *Builder) WithLogging(level, format string, accessLogDisabled bool) *Builder {
	if level != "" {
		_ = fb.With("configData.log.level", level)
		// restart deployment registry to fetch new logging configuration from configmap
		fb = fb.withRollme(fmt.Sprintf("configData.log.level=%s", level))
	}
	if format != "" {
		_ = fb.With("configData.log.formatter", format)
		// restart deployment registry to fetch new logging configuration from configmap
		fb = fb.withRollme(fmt.Sprintf("configData.log.formatter=%s", format))
	}
	// Access logs use Apache Combined Log Format and cannot use json/text formatter
	_ = fb.With("configData.log.accesslog.disabled", accessLogDisabled)
	fb = fb.withRollme(fmt.Sprintf("configData.log.accesslog.disabled=%t", accessLogDisabled))
	return fb
}

// withRollme allows to set custom values for the `rollme` field in chart
// it merges values for many command executions in format <value1>,<value2>,...,<valueN>
func (fb *Builder) withRollme(value string) *Builder {
	fb.rollmeValues = append(fb.rollmeValues, value)
	_ = fb.With("rollme", strings.Join(fb.rollmeValues, "\\,"))
	return fb
}
