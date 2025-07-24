provider "google" {
  project = "xxxxxxxxxxxxxxxxx"
  region  = "xxxxxxxxxxx"
  zone    = "xxxxxxxxxxxxx"
}

resource "doppler_integration_external_id" "i_gcp_sak_extid" {
  integration_type = "gcp_service_account_keys"
}

resource "google_service_account" "gcpsa_doppler_impersonate" {
  account_id   = "xxxxxxxxxxx"
  description  = format("doppler_impersonate:%s", doppler_integration_external_id.i_gcp_sak_extid.id)
  display_name = "xxxxxxxxxx"
}

data "google_iam_policy" "doppler_rotation_impersonate_policy" {
  binding {
    role = "roles/iam.serviceAccountViewer"
    members = [
      google_service_account.gcpsa_doppler_impersonate.member,
      "serviceAccount:operator@doppler-integrations.iam.gserviceaccount.com"
    ]
  }
  binding {
    role = "roles/iam.serviceAccountTokenCreator"
    members = [
      "serviceAccount:operator@doppler-integrations.iam.gserviceaccount.com"
    ]
  }
}

resource "google_service_account_iam_policy" "gcp_policy_impersonate" {
  service_account_id = google_service_account.gcpsa_doppler_impersonate.name
  policy_data        = data.google_iam_policy.doppler_rotation_impersonate_policy.policy_data
}

resource "google_service_account" "gcpsa_rotated" {
  account_id   = "xxxxxxxxxxxxxxxxxx"
  display_name = "xxxxxxxxxxxxxxxxxx"
}

data "google_iam_policy" "doppler_rotated_user_policy" {
  binding {
    role = "roles/iam.serviceAccountKeyAdmin"
    members = [
      google_service_account.gcpsa_doppler_impersonate.member,
    ]
  }
}

resource "google_service_account_iam_policy" "gcp_policy_rotated" {
  service_account_id = google_service_account.gcpsa_rotated.name
  policy_data        = data.google_iam_policy.doppler_rotated_user_policy.policy_data
}

resource "doppler_integration_gcp_service_account_keys" "i_gcp_sak" {
  name                         = "TF GCP Service Account Keys"
  impersonated_service_account = google_service_account.gcpsa_doppler_impersonate.email
  external_id                  = doppler_integration_external_id.i_gcp_sak_extid.id
}

resource "doppler_rotated_secret_gcp_service_account_keys" "rs_gcpsak" {
  integration         = doppler_integration_gcp_service_account_keys.i_gcp_sak.id
  project             = "backend"
  config              = "dev"
  name                = "GCP_SERVICE_ACCOUNT"
  rotation_period_sec = 2592000
  service_account     = google_service_account.gcpsa_rotated.email
}

