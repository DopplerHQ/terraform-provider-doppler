resource "doppler_trusted_ips" "backend_prd" {
  project = "backend"
  config  = "prd"
  trusted_ips = [
    "203.0.113.0/24",
    "198.51.100.5/32",
    "2001:db8::1/128"
  ]
}

resource "doppler_trusted_ips" "backend_stg" {
  project = "backend"
  config  = "stg"
  trusted_ips = ["203.0.113.0/24"]
}

resource "doppler_trusted_ips" "backend_dev" {
  project = "backend"
  config  = "dev"
  trusted_ips = ["0.0.0.0/0"]
}
