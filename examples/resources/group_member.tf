resource "doppler_group" "engineering" {
  name = "engineering"
}

data "doppler_user" "nic" {
  email = "nic@doppler.com"
}

data "doppler_user" "andre" {
  email = "andre@doppler.com"
}

resource "doppler_group_member" "engineering" {
  for_each = toset([data.doppler_user.nic.id, data.doppler_user.andre.id])
  group    = doppler_group.engineering.id
  user     = each.value
}