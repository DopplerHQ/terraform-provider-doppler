resource "doppler_secret_note" "stripe_key" {
  project = "backend"
  name    = "STRIPE_SECRET_KEY"
  note    = "Owner: payments team. Rotate quarterly via Stripe dashboard."
}
