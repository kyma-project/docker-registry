# Testing Strategy

## CI/CD Jobs Running on Pull Requests

Each pull request to the repository triggers the following CI/CD jobs that verify the Docker Registry Operator reconciliation logic and run integration tests of the Docker Registry module:

- `lint / lint` - Is responsible for the Operator linting and static code analysis. For the configuration, see the [lint.yaml](https://github.com/kyma-project/docker-registry/blob/main/.github/workflows/lint.yaml) file.
- `pull / unit-tests / unit-tests` - Runs basic unit tests of Operator's logic. For the configuration, see the [_unit-tests.yaml](https://github.com/kyma-project/docker-registry/blob/main/.github/workflows/_unit-tests.yaml) file.
- `pull / integrations / operator-integration-test` - Runs the create/update/delete Docker Registry integration tests in k3d cluster. For the configuration, see the [_integration-tests-pull.yaml](https://github.com/kyma-project/docker-registry/blob/main/.github/workflows/_integration-tests.yaml) file.

## CI/CD Jobs Running on the Main Branch

- `push / integrations / operator-integration-test` - Runs the create/update/delete Docker Registry integration tests in k3d cluster. For the configuration, see the [_integration-tests-push.yaml](https://github.com/kyma-project/docker-registry/blob/main/.github/workflows/_integration-tests.yaml) file.
- `push / integrations / gardener-integration-test` - Checks the installation of the Docker Registry module in the Gardener shoot cluster and runs basic integration tests of Docker Registry. For the configuration, see the [_integration-tests-push.yaml](https://github.com/kyma-project/docker-registry/blob/main/.github/workflows/_integration-tests.yaml) file.
- `push / upgrades / operator-upgrade-test` - Runs the upgrade integration test suite and verifies if the latest release can be successfully upgraded to the new (`main`) revision. For the configuration, see the [_upgrade-tests.yaml](https://github.com/kyma-project/docker-registry/blob/main/.github/workflows/_upgrade-tests.yaml) file.
- `markdown / link-check` - Checks if there are no broken links in `.md` files. For the configuration, see the [mlc.config.json](https://github.com/kyma-project/docker-registry/blob/main/.mlc.config.json) and the [markdown.yaml](https://github.com/kyma-project/docker-registry/blob/main/.github/workflows/markdown.yaml) files.

## CI/CD Jobs Running on a Schedule

- `markdown / link-check` - Runs Markdown link check every day at 05:00 AM. For the configuration, see the [mlc.config.json](https://github.com/kyma-project/docker-registry/blob/main/.mlc.config.json) and the [markdown.yaml](https://github.com/kyma-project/docker-registry/blob/main/.github/workflows/markdown.yaml) files.
