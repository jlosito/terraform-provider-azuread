package serviceprincipals

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-azuread/internal/clients"
	"github.com/terraform-providers/terraform-provider-azuread/internal/tf"
)

func clientConfigDataSourceReadMsGraph(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	objectId := ""
	if client.Claims.ObjectId != "" {
		objectId = client.Claims.ObjectId
	}

	d.SetId(fmt.Sprintf("%s-%s-%s", client.TenantID, client.ClientID, objectId))

	if dg := tf.Set(d, "tenant_id", client.TenantID); dg != nil {
		return dg
	}

	if dg := tf.Set(d, "client_id", client.ClientID); dg != nil {
		return dg
	}

	if dg := tf.Set(d, "object_id", objectId); dg != nil {
		return dg
	}

	return nil
}
