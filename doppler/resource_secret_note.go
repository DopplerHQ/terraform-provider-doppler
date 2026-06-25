package doppler

import (
	"context"
	"errors"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSecretNote() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSecretNoteUpsert,
		ReadContext:   resourceSecretNoteRead,
		UpdateContext: resourceSecretNoteUpsert,
		DeleteContext: resourceSecretNoteDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"project": {
				Description: "The name of the Doppler project that owns the secret",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "The name of the secret. Notes are scoped per project + secret name and apply across all configs.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"note": {
				Description: "The note content to attach to the secret",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func getSecretNoteId(project, name string) string {
	return strings.Join([]string{project, name}, ".")
}

func parseSecretNoteId(id string) (project string, name string, err error) {
	tokens := strings.SplitN(id, ".", 2)
	if len(tokens) != 2 {
		return "", "", errors.New("invalid secret note ID (expected `project.name`)")
	}
	return tokens[0], tokens[1], nil
}

func resourceSecretNoteUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	project := d.Get("project").(string)
	name := d.Get("name").(string)
	note := d.Get("note").(string)

	if err := client.SetSecretNote(ctx, project, name, note); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(getSecretNoteId(project, name))
	return resourceSecretNoteRead(ctx, d, m)
}

func resourceSecretNoteRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	project, name, err := parseSecretNoteId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	note, err := client.GetSecretNote(ctx, project, name)
	if err != nil {
		return handleNotFoundError(err, d)
	}

	if err := d.Set("project", project); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("note", note); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceSecretNoteDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	project, name, err := parseSecretNoteId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Clear the note by writing an empty string. The Doppler API has no
	// dedicated DELETE for secret notes; an empty note is the "unset" state.
	if err := client.SetSecretNote(ctx, project, name, ""); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
