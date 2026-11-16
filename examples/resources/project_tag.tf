resource "doppler_tag" "backend" {
  name = "Backend"
}

resource "doppler_project_tag" "backend" {
  project  = "backend"
  tag_slug = doppler_tag.backend.slug
}
