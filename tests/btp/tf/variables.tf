# we're using uppercase variable names, since in some cases (e.g Azure DevOps) the system variables are forced to be uppercase
# TF allows providing variable values as env variables of name name, case sensitive

variable "BTP_GLOBAL_ACCOUNT" {
  type        = string
  description = "Global account name"
  default     = "global-account-guid"
}

variable "BTP_BOT_USER" {
  type        = string
  description = "Bot account name"
  default     = "email@domain.com"
}

variable "BTP_BOT_PASSWORD" {
  type        = string
  description = "Bot account password"
  default     = "password"
}

variable "BTP_BACKEND_URL" {
  type        = string
  description = "BTP backend URL"
  default     = "https://cpcli.cf.sap.hana.ondemand.com"
}

variable "BTP_SUBACCOUNT" {
  type        = string
  description = "Subaccount name"
  default     = "subaccount-name"
}

variable "BTP_KYMA_PLAN" {
  type        = string
  description = "Plan name"
  default     = "gcp"
}

variable "BTP_SA_REGION" {
  type        = string
  description = "Region name"
  default     = "us31"
}

variable "BTP_CUSTOM_IAS_TENANT" {
  type        = string
  description = "Custom IAS tenant"
  default     = "custon-tenant"
}

variable "BTP_CUSTOM_IAS_DOMAIN" {
  type        = string
  description = "Custom IAS domain"
  default     = "accounts400.ondemand.com"
}

variable "BTP_KYMA_REGION" {
  type        = string
  description = "Kyma region"
  default     = "us-central1"
}

variable "BTP_PROVIDER_SUBACCOUNT_ID" {
  type        = string
  description = "Subaccount ID"
  default     = "subaccount-id"
}
