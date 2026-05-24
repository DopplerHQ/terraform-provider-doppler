data "doppler_project" "example" {
  name = "my-project"
}

output "project_slug" {
  value = data.doppler_project.example.slug
}

output "project_description" {
  value = data.doppler_project.example.description
}
