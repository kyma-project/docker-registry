

resource "btp_subaccount_entitlement" "kyma" {
  subaccount_id = btp_subaccount.subaccount.id
  service_name  = "kymaruntime"
  plan_name     = var.BTP_KYMA_PLAN
  amount        = 1
}

resource "btp_subaccount_environment_instance" "kyma" {
  subaccount_id    = btp_subaccount.subaccount.id
  name             = "${var.BTP_SUBACCOUNT}-kyma"
  environment_type = "kyma"
  service_name     = btp_subaccount_entitlement.kyma.service_name
  plan_name        = btp_subaccount_entitlement.kyma.plan_name
  parameters = jsonencode({
    modules = {
      list = [
        {
          name    = "api-gateway"
          channel = "fast"
        },
        {
          name    = "istio"
          channel = "fast"
        },
        {
          name    = "btp-operator"
          channel = "fast"
        }
      ]
    }
    oidc = {
      groupsClaim    = "groups"
      signingAlgs    = ["RS256"]
      usernameClaim  = "sub"
      usernamePrefix = "-"
      clientID       = jsondecode(btp_subaccount_service_binding.identity_application_binding.credentials).clientid
      issuerURL      = "https://${var.BTP_CUSTOM_IAS_TENANT}.${var.BTP_CUSTOM_IAS_DOMAIN}"
    }
    name   = "${var.BTP_SUBACCOUNT}-kyma"
    region = var.BTP_KYMA_REGION
    administrators = [
      var.BTP_BOT_USER
    ]
  })
  timeouts = {
    create = "40m"
    update = "30m"
    delete = "60m"
  }
}


resource "local_file" "kubeconfig_url" {
  content  = jsondecode(btp_subaccount_environment_instance.kyma.labels).KubeconfigURL
  filename = "kubeconfig_url.txt"
}
