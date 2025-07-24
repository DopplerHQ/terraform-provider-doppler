data "doppler_environments" "all" {
  project = "my-project"
}

output "environment_names" {
  value = [for env in data.doppler_environments.all.list : env.name]
}

output "environment_slugs" {
  value = [for env in data.doppler_environments.all.list : env.slug]
} 