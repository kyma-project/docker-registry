resource "btp_subaccount_entitlement" "identity" {
  subaccount_id = btp_subaccount.subaccount.id
  service_name  = "identity"
  plan_name     = "application"
}

# custom idp
resource "btp_subaccount_trust_configuration" "custom_idp" {
  subaccount_id     = btp_subaccount.subaccount.id
  identity_provider = "${var.BTP_CUSTOM_IAS_TENANT}.${var.BTP_CUSTOM_IAS_DOMAIN}"
  name              = "${var.BTP_SUBACCOUNT}-${var.BTP_CUSTOM_IAS_TENANT}"
  depends_on        = [btp_subaccount_entitlement.identity]
}

data "btp_subaccount_service_plan" "identity_application" {
  depends_on    = [btp_subaccount_entitlement.identity]
  subaccount_id = btp_subaccount.subaccount.id
  offering_name = "identity"
  name          = "application"
}

resource "btp_subaccount_service_instance" "identity_application" {
  depends_on     = [btp_subaccount_trust_configuration.custom_idp]
  subaccount_id  = btp_subaccount.subaccount.id
  name           = "${var.BTP_SUBACCOUNT}-${var.BTP_CUSTOM_IAS_TENANT}-oidc-app"
  serviceplan_id = data.btp_subaccount_service_plan.identity_application.id
  parameters = jsonencode({
    user-access = "public"
    oauth2-configuration = {
      grant-types = [
        "authorization_code",
        "authorization_code_pkce_s256",
        "password",
        "refresh_token"
      ],
      token-policy = {
        token-validity              = 3600,
        refresh-validity            = 15552000,
        refresh-usage-after-renewal = "off",
        refresh-parallel            = 3,
        access-token-format         = "default"
      },
      public-client = true,
      redirect-uris = [
        "https://dashboard.kyma.cloud.sap",
        "https://dashboard.dev.kyma.cloud.sap",
        "https://dashboard.stage.kyma.cloud.sap",
        "http://localhost:8000"
      ]
    },
    subject-name-identifier = {
      attribute          = "mail",
      fallback-attribute = "none"
    },
    default-attributes = null,
    assertion-attributes = {
      email      = "mail",
      groups     = "companyGroups",
      first_name = "firstName",
      last_name  = "lastName",
      login_name = "loginName",
      mail       = "mail",
      scope      = "companyGroups",
      user_uuid  = "userUuid",
      locale     = "language"
    },
    name         = "${var.BTP_SUBACCOUNT}-${var.BTP_CUSTOM_IAS_TENANT}-oidc-app",
    display-name = "${var.BTP_SUBACCOUNT}-${var.BTP_CUSTOM_IAS_TENANT}-oidc-app"
  })
}

resource "btp_subaccount_service_binding" "identity_application_binding" {
  subaccount_id       = btp_subaccount.subaccount.id
  name                = "${var.BTP_SUBACCOUNT}-${var.BTP_CUSTOM_IAS_TENANT}-oidc-app-binding"
  service_instance_id = btp_subaccount_service_instance.identity_application.id
  parameters = jsonencode({
    credential-type = "X509_GENERATED"
    key-length      = 4096
    validity        = 1
    validity-type   = "DAYS"
    app-identifier  = "kymaruntime"
  })
}

locals {
  identity_credentials = jsondecode(btp_subaccount_service_binding.identity_application_binding.credentials)
}

resource "local_file" "binding_credentials" {
  content = jsonencode({
    clientid = local.identity_credentials.clientid
    url      = local.identity_credentials.url
  })
  filename = "binding_credentials.json"
}
