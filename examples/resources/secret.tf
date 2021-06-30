resource "random_password" "db_password" {
  length = 32
  special = true
}

resource "doppler_secret" "db_password" {
  project = "backend"
  config = "dev"
  name = "DB_PASSWORD"
  value = random_password.db_password.result
}

output "resource_value" {
  # Access the secret value
  # nonsensitive used for demo purposes only
  value = nonsensitive(doppler_secret.db_password.value)
}
