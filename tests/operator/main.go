package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/kyma-project/docker-registry/components/operator/api/v1alpha1"
	"github.com/kyma-project/docker-registry/tests/operator/dockerregistry"
	"github.com/kyma-project/docker-registry/tests/operator/logger"
	"github.com/kyma-project/docker-registry/tests/operator/namespace"
	"github.com/kyma-project/docker-registry/tests/operator/utils"
)

var (
	testTimeout = time.Minute * 10
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
