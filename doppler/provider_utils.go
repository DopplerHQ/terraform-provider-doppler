package doppler

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type CustomNotFoundError struct {
	Message string
}

func (e *CustomNotFoundError) Error() string {
	return fmt.Sprintf("Doppler Error: %s", e.Message)
}

func handleNotFoundError(err error, d *schema.ResourceData) diag.Diagnostics {
	isNotFoundError := false

	if apiError, ok := err.(*APIError); ok && apiError.Response != nil && apiError.Response.HTTPResponse.StatusCode == 404 {
		isNotFoundError = true
	}

	if _, ok := err.(*CustomNotFoundError); ok {
		isNotFoundError = true
	}

	if isNotFoundError {
		// the resource no longer exists, so reset its ID so Terraform will
		// generate a plan that recreates it
		d.SetId("")
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  err.Error(),
				Detail:   "Resource was not found, so it was removed from state and is being recreated.",
			},
		}
	}

	return diag.FromErr(err)
}
