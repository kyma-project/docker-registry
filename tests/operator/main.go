package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/tests/operator/dockerregistry"
	"github.com/kyma-project/docker-registry/tests/operator/dockerregistry/deployment"
	"github.com/kyma-project/docker-registry/tests/operator/logger"
	"github.com/kyma-project/docker-registry/tests/operator/namespace"
	"github.com/kyma-project/docker-registry/tests/operator/secret"
	"github.com/kyma-project/docker-registry/tests/operator/utils"
)

const storageSecretName = "storage-credentials"

var (
	testTimeout = time.Minute * 15

	storageCredentials = map[string]string{
		"accessKey": "initial-access-key",
		"secretKey": "initial-secret-key",
	}
	rotatedStorageCredentials = map[string]string{
		"accessKey": "rotated-access-key",
		"secretKey": "rotated-secret-key",
	}
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	log, err := logger.New()
	if err != nil {
		fmt.Printf("%s: %s\n", "unable to setup logger", err)
		os.Exit(1)
	}

	log.Info("Configuring test essentials")
	client, err := utils.GetKuberentesClient()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	log.Info("Start scenario")
	err = runScenario(&utils.TestUtils{
		Namespace: fmt.Sprintf("dockerregistry-test-%s", uuid.New().String()),
		Ctx:       ctx,
		Client:    client,
		Logger:    log,

		Name:                     "default-test",
		DockerregistryDeployName: "dockerregistry",
		RegistryName:             "dockerregistry-docker-registry",
		UpdateSpec:               v1alpha1.DockerRegistrySpec{},
	})
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	// The storage secret scenario asserts that the operator recovers on its own once a
	// missing storage Secret shows up. That self-recovery relies on the storage retry
	// requeue (introduced in #580) and the Warning configuration flow (#587). The upgrade
	// test runs this same binary twice: first against the latest released operator, then
	// against the freshly built one. The released operator predates both changes and can
	// never recover, so the scenario would hang until the test times out. Gate it behind
	// an env var that the workflow sets only for the run against the new operator.
	if os.Getenv("RUN_STORAGE_SECRET_SCENARIO") != "true" {
		log.Info("Skipping storage secret scenario (RUN_STORAGE_SECRET_SCENARIO != true)")
		return
	}

	log.Info("Start storage secret scenario")
	err = runStorageSecretScenario(&utils.TestUtils{
		Namespace: fmt.Sprintf("dockerregistry-test-%s", uuid.New().String()),
		Ctx:       ctx,
		Client:    client,
		Logger:    log,

		Name:                     "storage-secret-test",
		DockerregistryDeployName: "dockerregistry",
		RegistryName:             "dockerregistry-docker-registry",
		StorageSecretName:        storageSecretName,
		StorageSecretData:        storageCredentials,
		// the bucket is never reached, the scenario only checks how the operator reacts to the Secret
		CreateSpec: v1alpha1.DockerRegistrySpec{
			Storage: &v1alpha1.Storage{
				S3: &v1alpha1.StorageS3{
					Bucket:         "test-bucket",
					Region:         "us-east-1",
					RegionEndpoint: "http://storage.local:9000",
					SecretName:     storageSecretName,
				},
			},
		},
	})
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func runScenario(testutil *utils.TestUtils) error {
	// create test namespace
	testutil.Logger.Infof("Creating namespace '%s'", testutil.Namespace)
	if err := namespace.Create(testutil); err != nil {
		return err
	}

	// create Docker Registry
	testutil.Logger.Infof("Creating dockerregistry '%s'", testutil.Name)
	if err := dockerregistry.Create(testutil); err != nil {
		return err
	}

	// verify Docker Registry
	testutil.Logger.Infof("Verifying dockerregistry '%s'", testutil.Name)
	if err := utils.WithRetry(testutil, dockerregistry.Verify); err != nil {
		return err
	}

	// update Docker Registry with other spec
	testutil.Logger.Infof("Updating dockerregistry '%s'", testutil.Name)
	if err := dockerregistry.Update(testutil); err != nil {
		return err
	}

	// verify Docker Registry
	testutil.Logger.Infof("Verifying dockerregistry '%s'", testutil.Name)
	if err := utils.WithRetry(testutil, dockerregistry.Verify); err != nil {
		return err
	}

	// delete Docker Registry
	testutil.Logger.Infof("Deleting dockerregistry '%s'", testutil.Name)
	if err := dockerregistry.Delete(testutil); err != nil {
		return err
	}

	// verify Docker Registry deletion
	testutil.Logger.Infof("Verifying dockerregistry '%s' deletion", testutil.Name)
	if err := utils.WithRetry(testutil, dockerregistry.VerifyDeletion); err != nil {
		return err
	}

	// cleanup
	testutil.Logger.Infof("Deleting namespace '%s'", testutil.Namespace)
	return namespace.Delete(testutil)
}

// runStorageSecretScenario covers a Docker Registry backed by external storage whose Secret shows up late
// and is rotated later on. Both flows have to work without touching the DockerRegistry CR.
func runStorageSecretScenario(testutil *utils.TestUtils) error {
	// create test namespace
	testutil.Logger.Infof("Creating namespace '%s'", testutil.Namespace)
	if err := namespace.Create(testutil); err != nil {
		return err
	}

	// create Docker Registry referencing a storage Secret that does not exist yet
	testutil.Logger.Infof("Creating dockerregistry '%s' without the storage secret '%s'",
		testutil.Name, testutil.StorageSecretName)
	if err := dockerregistry.Create(testutil); err != nil {
		return err
	}

	// verify Docker Registry reports the missing Secret
	testutil.Logger.Infof("Verifying dockerregistry '%s' warns about the missing storage secret", testutil.Name)
	if err := utils.WithRetry(testutil, dockerregistry.VerifyMissingStorageSecretWarning); err != nil {
		return err
	}

	// create the missing storage Secret
	testutil.Logger.Infof("Creating storage secret '%s'", testutil.StorageSecretName)
	if err := secret.Create(testutil); err != nil {
		return err
	}

	// verify Docker Registry recovers on its own
	testutil.Logger.Infof("Verifying dockerregistry '%s' recovers from the warning", testutil.Name)
	if err := utils.WithRetry(testutil, dockerregistry.Verify); err != nil {
		return err
	}
	if err := utils.WithRetry(testutil, dockerregistry.VerifyStorageCredentials); err != nil {
		return err
	}

	podBeforeRotation, err := deployment.RegistryPodName(testutil)
	if err != nil {
		return err
	}

	// rotate the credentials in the storage Secret
	testutil.Logger.Infof("Rotating credentials in the storage secret '%s'", testutil.StorageSecretName)
	testutil.StorageSecretData = rotatedStorageCredentials
	if err := secret.Update(testutil); err != nil {
		return err
	}

	// verify the registry runs with the rotated credentials
	testutil.Logger.Infof("Verifying dockerregistry '%s' restarts with the rotated credentials", testutil.Name)
	if err := utils.WithRetry(testutil, func(testutil *utils.TestUtils) error {
		return deployment.VerifyRegistryRestarted(testutil, podBeforeRotation)
	}); err != nil {
		return err
	}
	if err := utils.WithRetry(testutil, dockerregistry.VerifyStorageCredentials); err != nil {
		return err
	}
	if err := utils.WithRetry(testutil, dockerregistry.Verify); err != nil {
		return err
	}

	// delete Docker Registry
	testutil.Logger.Infof("Deleting dockerregistry '%s'", testutil.Name)
	if err := dockerregistry.Delete(testutil); err != nil {
		return err
	}

	// verify Docker Registry deletion
	testutil.Logger.Infof("Verifying dockerregistry '%s' deletion", testutil.Name)
	if err := utils.WithRetry(testutil, dockerregistry.VerifyDeletion); err != nil {
		return err
	}

	// cleanup
	testutil.Logger.Infof("Deleting namespace '%s'", testutil.Namespace)
	return namespace.Delete(testutil)
}
