package groups

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/manicminer/hamilton/models"

	"github.com/terraform-providers/terraform-provider-azuread/internal/clients"
	"github.com/terraform-providers/terraform-provider-azuread/internal/tf"
)

func groupDataSourceReadMsGraph(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client).Groups.MsClient

	var group models.Group
	var displayName string

	if v, ok := d.GetOk("display_name"); ok {
		displayName = v.(string)
	} else if v, ok := d.GetOk("name"); ok {
		displayName = v.(string)
	}

	if displayName != "" {
		filter := fmt.Sprintf("displayName eq '%s'", displayName)
		groups, _, err := client.List(ctx, filter)
		if err != nil {
			return tf.ErrorDiagPathF(err, "name", "No group found with display name: %q", displayName)
		}

		count := len(*groups)
		if count > 1 {
			return tf.ErrorDiagPathF(nil, "name", "More than one group found with display name: %q", displayName)
		} else if count == 0 {
			return tf.ErrorDiagPathF(err, "name", "No group found with display name: %q", displayName)
		}

		group = (*groups)[0]
	} else if objectId, ok := d.Get("object_id").(string); ok && objectId != "" {
		g, status, err := client.Get(ctx, objectId)
		if err != nil {
			if status == http.StatusNotFound {
				return tf.ErrorDiagPathF(nil, "object_id", "No group found with object ID: %q", objectId)
			}
			return tf.ErrorDiagPathF(err, "object_id", "Retrieving group with object ID: %q", objectId)
		}
		group = *g
	}

	if group.ID == nil {
		return tf.ErrorDiagF(errors.New("API returned group with nil object ID"), "Bad API Response")
	}

	d.SetId(*group.ID)

	if dg := tf.Set(d, "object_id", group.ID); dg != nil {
		return dg
	}

	if dg := tf.Set(d, "display_name", group.DisplayName); dg != nil {
		return dg
	}

	// TODO: v2.0 remove this
	if dg := tf.Set(d, "name", group.DisplayName); dg != nil {
		return dg
	}

	if dg := tf.Set(d, "description", group.Description); dg != nil {
		return dg
	}

	members, _, err := client.ListMembers(ctx, d.Id())
	if err != nil {
		return tf.ErrorDiagF(err, "Could not retrieve group members for group with object ID: %q", d.Id())
	}

	if dg := tf.Set(d, "members", members); dg != nil {
		return dg
	}

	owners, _, err := client.ListOwners(ctx, d.Id())
	if err != nil {
		return tf.ErrorDiagF(err, "Could not retrieve group owners for group with object ID: %q", d.Id())
	}

	if dg := tf.Set(d, "owners", owners); dg != nil {
		return dg
	}

	return nil
}
