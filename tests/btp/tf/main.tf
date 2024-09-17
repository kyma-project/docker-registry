terraform {
  required_providers {
    btp = {
      source  = "SAP/btp"
      version = "1.6.0"
    }
    jq = {
      source  = "massdriver-cloud/jq"
    }
    http = {
      source = "hashicorp/http"
      version = "3.4.5"
    }
  }
}

provider "jq" {}
provider "http" {}

provider "btp" {
  globalaccount = var.BTP_GLOBAL_ACCOUNT
  cli_server_url = var.BTP_BACKEND_URL
  idp            = var.BTP_CUSTOM_IAS_TENANT
  username = var.BTP_BOT_USER
  password = var.BTP_BOT_PASSWORD
}

module "kyma" {
  source = "github.com/kyma-project/terraform-module"
  BTP_NEW_SUBACCOUNT_NAME = var.BTP_NEW_SUBACCOUNT_NAME
  BTP_CUSTOM_IAS_TENANT = var.BTP_CUSTOM_IAS_TENANT
  BTP_BOT_USER = var.BTP_BOT_USER
  BTP_BOT_PASSWORD = var.BTP_BOT_PASSWORD
  BTP_PROVIDER_SUBACCOUNT_ID = var.BTP_PROVIDER_SUBACCOUNT_ID
}

resource "local_file" "provider_sm" {
  content  = <<EOT
clientid=${module.kyma.custom_service_manager_credentials.clientid}
clientsecret=${module.kyma.custom_service_manager_credentials.clientsecret}
sm_url=${module.kyma.custom_service_manager_credentials.sm_url}
tokenurl=${module.kyma.custom_service_manager_credentials.url}
tokenurlsuffix=/oauth/token
EOT
  filename = "provider-sm-decoded.env"
}


output "subaccount_id" {
  value = module.kyma.subaccount_id
}
