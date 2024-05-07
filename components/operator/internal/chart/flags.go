package chart

import (
	"fmt"
	"strings"
)

type FlagsBuilder interface {
	Build() map[string]interface{}
	WithControllerConfiguration(healthzLivenessTimeout string) *flagsBuilder
	WithRegistryCredentials(username string, password string) *flagsBuilder
	WithRegistryHttpSecret(httpSecret string) *flagsBuilder
	WithNodePort(nodePort int64) *flagsBuilder
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
