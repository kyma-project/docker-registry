resource "btp_subaccount" "subaccount" {
  name      = var.BTP_SUBACCOUNT
  region    = var.BTP_SA_REGION
  subdomain = var.BTP_SUBACCOUNT
}

resource "local_file" "subaccount_id" {
  content  = btp_subaccount.subaccount.id
  filename = "subaccount_id.txt"
}
