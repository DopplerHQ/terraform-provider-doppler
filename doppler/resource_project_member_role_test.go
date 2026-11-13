package doppler

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviderFactories = map[string]func() (*schema.Provider, error){
	"doppler": func() (*schema.Provider, error) {
		return Provider(), nil
	},
}

func testAccPreCheck(t *testing.T) {
	if os.Getenv("DOPPLER_TOKEN") == "" {
		t.Fatal("DOPPLER_TOKEN must be set for acceptance tests")
	}
}

func testAccProjectMemberServiceAccountRoleAssignmentConfig(name string, configs string) string {
	return fmt.Sprintf(`
resource "doppler_project" "test" {
  name = %[1]q
}

resource "doppler_environment" "prd" {
  project = doppler_project.test.name
  slug    = "prd"
  name    = "Production"
}

resource "doppler_config" "prd_a" {
  project     = doppler_project.test.name
  environment = doppler_environment.prd.slug
  name        = "prd_a"
}

resource "doppler_config" "prd_b" {
  project     = doppler_project.test.name
  environment = doppler_environment.prd.slug
  name        = "prd_b"
}

resource "doppler_service_account" "test" {
  name           = %[1]q
  workplace_role = "no_access"
}

resource "doppler_project_member_service_account_role_assignment" "test" {
  project              = doppler_project.test.name
  service_account_slug = doppler_service_account.test.slug
  role                 = "viewer"
  configs              = %[2]s
}
`, name, configs)
}

func TestAccProjectMemberServiceAccountRoleAssignment_configs(t *testing.T) {
	name := fmt.Sprintf("tf-acc-%d", time.Now().UnixNano())
	resourceName := "doppler_project_member_service_account_role_assignment.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectMemberServiceAccountRoleAssignmentConfig(name, `[doppler_config.prd_a.name]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "environments.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "configs.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "configs.*", "prd_a"),
				),
			},
			{
				Config: testAccProjectMemberServiceAccountRoleAssignmentConfig(name, `[doppler_config.prd_b.name, doppler_environment.prd.slug]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "environments.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "configs.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "configs.*", "prd_b"),
					resource.TestCheckTypeSetElemAttr(resourceName, "configs.*", "prd"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
