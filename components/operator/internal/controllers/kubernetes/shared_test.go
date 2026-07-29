package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/labels"
)

func TestCredentialsSecretSelector(t *testing.T) {
	testCases := map[string]struct {
		labels  map[string]string
		matches bool
	}{
		"matches the propagated credentials secret": {
			labels:  map[string]string{ConfigLabel: CredentialsLabelValue},
			matches: true,
		},
		"matches a credentials secret with additional labels": {
			labels:  map[string]string{ConfigLabel: CredentialsLabelValue, "app": "dockerregistry"},
			matches: true,
		},
		"skips another config secret": {
			labels:  map[string]string{ConfigLabel: "something-else"},
			matches: false,
		},
		// storage credentials are created by users, they never carry operator labels, and caching
		// them would make the operator memory follow the number of Secrets in the cluster
		"skips a storage secret": {
			labels:  map[string]string{"app": "storage"},
			matches: false,
		},
		"skips a secret without labels": {
			labels:  nil,
			matches: false,
		},
	}

	selector := CredentialsSecretSelector()

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, testCase.matches, selector.Matches(labels.Set(testCase.labels)))
		})
	}
}
