data "btp_subaccount_service_binding" "provider_sm" {
  subaccount_id = var.BTP_PROVIDER_SUBACCOUNT_ID
  name          = "provider-sm-binding"
}


locals {
  provider_credentials = jsondecode(data.btp_subaccount_service_binding.provider_sm.credentials)
}

resource "local_file" "provider_sm" {
  content  = <<EOT
clientid=${local.provider_credentials.clientid}
clientsecret=${local.provider_credentials.clientsecret}
sm_url=${local.provider_credentials.sm_url}
tokenurl=${local.provider_credentials.url}
tokenurlsuffix=/oauth/token
EOT
  filename = "provider-sm-decoded.env"
}
