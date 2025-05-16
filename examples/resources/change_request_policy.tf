data "doppler_user" "nic" {
  email = "nic@doppler.com"
}

data "doppler_user" "emily" {
  email = "emily@doppler.com"
}

resource "doppler_project" "test_proj" {
  name = "my-test-project"
  description = "This is a test project"
}

resource "doppler_environment" "prd" {
  project = doppler_project.test_proj.name
  slug = "prd"
  name = "prd"
}

resource "doppler_environment" "ci" {
  project = doppler_project.test_proj.name
  slug = "ci"
  name = "CI-CD"
}

resource "doppler_config" "ci_github" {
  project = doppler_project.test_proj.name
  environment = doppler_environment.ci.slug
  name = "ci_github"
}

resource "doppler_group" "prod_reviewers" {
  name = "Prod Reviewers"
}

resource "doppler_group_member" "prod_reviewers" {
  for_each   = toset([data.doppler_user.nic.slug])
  group_slug = doppler_group.prod_reviewers.slug
  user_slug  = each.value
}

resource "doppler_change_request_policy" "prd_review" {
  name = "Prod Review"
  description = <<EOT
A change request policy which requires 2 total approvals, and 1 approval from emily or any member of the prod_reviewers group.
Reviews from the author of the change request are not counted.
This policy is enforced in all configs of the prd environment of test_proj, as well as the ci_github branch config.
EOT
  rules {
    disallow_self_review = true
    auto_assign_reviewers = "matchCount"

    required_reviewers {
      count = 2
    }

    required_reviewers {
      count = 1
      group_slugs = [doppler_group.prod_reviewers.slug]
      user_slugs = [data.doppler_user.emily.slug]
    }
  }
  targets {
    project {
      project_name = doppler_project.test_proj.name
      environment_slugs = [doppler_environment.prd.slug]
      config_names = [doppler_config.ci_github.name]
    }
  }
}