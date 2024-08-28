# BTP integration test

## Overview

This integration test verifies if the Docker Registry works in a semi-production environment.

## How to use

Export the following environment variables:
```bash
TF_VAR_BTP_BOT_USER=
TF_VAR_BTP_BOT_PASSWORD=
TF_VAR_BTP_GLOBAL_ACCOUNT=
TF_VAR_BTP_BACKEND_URL=https://cpcli.cf.sap.hana.ondemand.com
TF_VAR_BTP_CUSTOM_IAS_TENANT=
TF_VAR_BTP_CUSTOM_IAS_DOMAIN=accounts400.ondemand.com
TF_VAR_BTP_KYMA_PLAN=gcp
TF_VAR_BTP_SA_REGION=us31
TF_VAR_BTP_KYMA_REGION=us-central1
TF_VAR_BTP_PROVIDER_SUBACCOUNT_ID=
```

You can use the following command to export variables from the `.env` file which contains the above variables:
```bash
export $(cat .env | xargs)
```

