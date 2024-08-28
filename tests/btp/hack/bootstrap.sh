#!/bin/bash
set -e

# Operating system architecture
OS_ARCH=$(uname -m)
# Operating system type
OS_TYPE=$(uname | tr '[:upper:]' '[:lower:]')

if [ -z "$1" ]
  then
    echo "subaccount name not provided"
	exit 1
fi

mkdir -p tmp

#ensure kyma CLI into /bin folder
if [ ! -f ../bin/kyma ]; then
    echo "Kyma binary not found!"
    mkdir -p ../bin
    curl -s -L "https://github.com/kyma-project/cli/releases/download/v0.0.0-dev/kyma_$(uname -s)_$(uname -m).tar.gz" | tar -zxvf - -C ../bin kyma
    echo "Kyma binary downloaded into /bin/kyma"
fi

if [ ! -f ../bin/btp ]; then
    BTP_FILE=btp-cli-${OS_TYPE}-${OS_ARCH}-latest.tar.gz
    echo "BTP CLI not found!"
    mkdir -p ../bin
    curl -LJO https://tools.hana.ondemand.com/additional/${BTP_FILE} --cookie "eula_3_2_agreed=tools.hana.ondemand.com/developer-license-3_2.txt"
    tar -zxf ${BTP_FILE} --strip-components=1 -C ../bin
    rm -f ${BTP_FILE}
    echo "BTP CLI downloaded into /bin/btp"
fi


### TODO refactor to fetching btp Access Token manually via curl towards trusted IAS tenant.

../bin/btp login --url $TF_VAR_BTP_BACKEND_URL --user $TF_VAR_BTP_BOT_USER --password $TF_VAR_BTP_BOT_PASSWORD --idp $TF_VAR_BTP_CUSTOM_IAS_TENANT --subdomain $TF_VAR_BTP_GLOBAL_ACCOUNT

../bin/btp set config --format json

export TF_VAR_BTP_SUBACCOUNT=$1

# Create a new subaccount with Kyma instance and OIDC
tofu -chdir=../tf init
tofu -chdir=../tf apply -auto-approve
../bin/btp  target -sa $(cat ../tf/subaccount_id.txt)

##### ---------------------------------------------------------------------------------


### TODO: refactor getting access to the cluster : use btp CLI or kyma CLI ---

##  `kyma alpha get access btp --btpAccessToken --kymaToken --output`  https://github.com/kyma-project/cli/issues/2198

#Generate bot user based access
make headless-kubeconfig

#Generate acces based on service account bound to a selected cluster-role (for the automation purpose) using the one-off bot user based access
CLUSTERROLE=cluster-admin make service-account-kubeconfig

### ---------------------------------------------------------------------------

# add bindings to statefull service instances provisioned in different subaccount (btp object store)
# TF_VAR_BTP_BOT_USER must be assigned to the `Subaccount_viewer` role collection in the provider subaccount (TF_VAR_BTP_PROVIDER_SUBACCOUNT_ID)
KUBECONFIG=tmp/sa-kubeconfig.yaml BTP_PROVIDER_SUBACCOUNT_ID=$TF_VAR_BTP_PROVIDER_SUBACCOUNT_ID make share-sm-service-operator-access
KUBECONFIG=tmp/sa-kubeconfig.yaml make create-object-store-reference

### ---------------------------------------------------------------------------

# TODO: change to enable from experimental channel via kyma v3 cli
KUBECONFIG=tmp/sa-kubeconfig.yaml make enable_docker_registry


# TODO: the following is sort of "kyma push app" equivalent for "cf push"
KUBECONFIG=tmp/sa-kubeconfig.yaml make docker_registry_login
KUBECONFIG=tmp/sa-kubeconfig.yaml make docker_push_simple_app
KUBECONFIG=tmp/sa-kubeconfig.yaml make deploy_simple_app


# CLEANUP
tofu -chdir=../tf destroy -auto-approve
rm -rf tmp
