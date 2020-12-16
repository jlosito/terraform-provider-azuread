package groups

import (
	"context"
	"errors"

	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-azuread/internal/clients"
	"github.com/terraform-providers/terraform-provider-azuread/internal/helpers/aadgraph"
	"github.com/terraform-providers/terraform-provider-azuread/internal/tf"
	"github.com/terraform-providers/terraform-provider-azuread/internal/utils"
)

func groupDataSourceReadAadGraph(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client).Groups.AadClient

	var group graphrbac.ADGroup
	var name string

	if v, ok := d.GetOk("display_name"); ok {
		name = v.(string)
	} else if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	}

	if objectId, ok := d.Get("object_id").(string); ok && objectId != "" {
		resp, err := client.Get(ctx, objectId)
		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return tf.ErrorDiagPathF(nil, "object_id", "No group found with object ID: %q", objectId)
			}

			return tf.ErrorDiagF(err, "Retrieving group with object ID: %q", objectId)
		}

		group = resp
	} else if name != "" {
		g, err := aadgraph.GroupGetByDisplayName(ctx, client, name)
		if err != nil {
			return tf.ErrorDiagPathF(err, "name", "No group found with display name: %q", name)
		}
		group = *g
	}

	if group.ObjectID == nil {
		return tf.ErrorDiagF(errors.New("API returned group with nil object ID"), "Bad API Response")
	}

	d.SetId(*group.ObjectID)

	if dg := tf.Set(d, "object_id", group.ObjectID); dg != nil {
		return dg
	}

	if dg := tf.Set(d, "display_name", group.DisplayName); dg != nil {
		return dg
	}

	if dg := tf.Set(d, "name", group.DisplayName); dg != nil {
		return dg
	}

	description := ""
	if v, ok := group.AdditionalProperties["description"]; ok {
		description = v.(string)
	}
	if dg := tf.Set(d, "description", description); dg != nil {
		return dg
	}

	members, err := aadgraph.GroupAllMembers(ctx, client, d.Id())
	if err != nil {
		return tf.ErrorDiagPathF(err, "owners", "Could not retrieve members for group with object ID %q", d.Id())
	}

	if dg := tf.Set(d, "members", members); dg != nil {
		return dg
	}

	owners, err := aadgraph.GroupAllOwners(ctx, client, d.Id())
	if err != nil {
		return tf.ErrorDiagPathF(err, "owners", "Could not retrieve owners for group with object ID %q", d.Id())
	}

	if dg := tf.Set(d, "owners", owners); dg != nil {
		return dg
	}

	return nil
}
