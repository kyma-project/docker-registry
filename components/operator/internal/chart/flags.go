package chart

import (
	"fmt"
	"strings"

	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
)

type FlagsBuilder interface {
	Build() map[string]interface{}
	WithControllerConfiguration(healthzLivenessTimeout string) *flagsBuilder
	WithRegistryCredentials(username string, password string) *flagsBuilder
	WithRegistryHttpSecret(httpSecret string) *flagsBuilder
	WithNodePort(nodePort int64) *flagsBuilder
	WithAzure(secret *v1alpha1.StorageAzureSecrets) *flagsBuilder
	WithS3(config *v1alpha1.StorageS3, secret *v1alpha1.StorageS3Secrets) *flagsBuilder
	WithFilesystem() *flagsBuilder
}

type flagsBuilder struct {
	flags map[string]interface{}
}

func NewFlagsBuilder() FlagsBuilder {
	return &flagsBuilder{
		flags: map[string]interface{}{},
	}
}

func (fb *flagsBuilder) Build() map[string]interface{} {
	flags := map[string]interface{}{}
	for key, value := range fb.flags {
		flagPath := strings.Split(key, ".")
		appendFlag(flags, flagPath, value)
	}
	return flags
}

func appendFlag(flags map[string]interface{}, flagPath []string, value interface{}) {
	currentFlag := flags
	for i, pathPart := range flagPath {
		createIfEmpty(currentFlag, pathPart)
		if lastElement(flagPath, i) {
			currentFlag[pathPart] = value
		} else {
			currentFlag = nextDeeperFlag(currentFlag, pathPart)
		}
	}
}

func createIfEmpty(flags map[string]interface{}, key string) {
	if _, ok := flags[key]; !ok {
		flags[key] = map[string]interface{}{}
	}
}

func lastElement(values []string, i int) bool {
	return i == len(values)-1
}

func nextDeeperFlag(currentFlag map[string]interface{}, path string) map[string]interface{} {
	return currentFlag[path].(map[string]interface{})
}

func (fb *flagsBuilder) WithControllerConfiguration(healthzLivenessTimeout string) *flagsBuilder {
	optionalFlags := []struct {
		key   string
		value string
	}{
		{"healthzLivenessTimeout", healthzLivenessTimeout},
	}

	for _, flag := range optionalFlags {
		if flag.value != "" {
			fullPath := fmt.Sprintf("containers.manager.configuration.data.%s", flag.key)
			fb.flags[fullPath] = flag.value
		}
	}

	return fb
}

func (fb *flagsBuilder) WithRegistryCredentials(username, password string) *flagsBuilder {
	fb.flags["dockerRegistry.username"] = username
	fb.flags["dockerRegistry.password"] = password
	return fb
}

func (fb *flagsBuilder) WithRegistryHttpSecret(httpSecret string) *flagsBuilder {
	fb.flags["rollme"] = "dontrollplease"
	fb.flags["registryHTTPSecret"] = httpSecret
	return fb
}

func (fb *flagsBuilder) WithNodePort(nodePort int64) *flagsBuilder {
	fb.flags["registryNodePort"] = nodePort
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

func (fb *flagsBuilder) WithFilesystem() *flagsBuilder {
	fb.flags["storage"] = "filesystem"
	fb.flags["configData.storage.filesystem.rootdirectory"] = "/var/lib/registry"
	return fb
}
