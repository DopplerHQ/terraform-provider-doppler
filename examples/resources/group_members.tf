resource "doppler_group" "engineering" {
  name = "engineering"
}

data "doppler_user" "nic" {
  email = "nic@doppler.com"
}

data "doppler_user" "andre" {
  email = "andre@doppler.com"
}

resource "doppler_group_members" "engineering" {
  group_slug = doppler_group.engineering.slug
  user_slugs = [
    data.doppler_user.nic.slug,
    data.doppler_user.andre.slug
  ]
}

