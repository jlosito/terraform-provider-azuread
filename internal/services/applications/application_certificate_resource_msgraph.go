package applications

import (
	"context"
	"github.com/manicminer/hamilton/models"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-azuread/internal/clients"
	"github.com/terraform-providers/terraform-provider-azuread/internal/helpers/msgraph"
	"github.com/terraform-providers/terraform-provider-azuread/internal/services/applications/parse"
	"github.com/terraform-providers/terraform-provider-azuread/internal/tf"
)

func applicationCertificateResourceCreateMsGraph(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client).Applications.MsClient
	objectId := d.Get("application_object_id").(string)

	credential, err := msgraph.KeyCredentialForResource(d)
	if err != nil {
		attr := ""
		if kerr, ok := err.(msgraph.CredentialError); ok {
			attr = kerr.Attr()
		}
		return tf.ErrorDiagPathF(err, attr, "Generating certificate credentials for application with object ID %q", objectId)
	}

	id := parse.NewCredentialID(objectId, "certificate", *credential.KeyId)

	tf.LockByName(resourceApplicationName, id.ObjectId)
	defer tf.UnlockByName(resourceApplicationName, id.ObjectId)

	app, status, err := client.Get(ctx, id.ObjectId)
	if err != nil {
		if status == http.StatusNotFound {
			return tf.ErrorDiagPathF(nil, "application_object_id", "Application with object ID %q was not found", id.ObjectId)
		}
		return tf.ErrorDiagPathF(err, "application_object_id", "Retrieving Application with object ID %q", id.ObjectId)
	}

	newCredentials := make([]models.KeyCredential, 0)
	if app.KeyCredentials != nil {
		for _, cred := range *app.KeyCredentials {
			if cred.KeyId != nil && *cred.KeyId == *credential.KeyId {
				return tf.ImportAsExistsDiag("azuread_application_certificate", id.String())
			}
			newCredentials = append(newCredentials, cred)
		}
	}

	newCredentials = append(newCredentials, *credential)

	properties := models.Application{
		ID:             &id.ObjectId,
		KeyCredentials: &newCredentials,
	}
	if _, err := client.Update(ctx, properties); err != nil {
		return tf.ErrorDiagF(err, "Adding certificate for application with object ID %q", id.ObjectId)
	}

	d.SetId(id.String())

	return applicationCertificateResourceReadMsGraph(ctx, d, meta)
}

func applicationCertificateResourceReadMsGraph(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client).Applications.MsClient

	id, err := parse.CertificateID(d.Id())
	if err != nil {
		return tf.ErrorDiagPathF(err, "id", "Parsing certificate credential with ID %q", d.Id())
	}

	app, status, err := client.Get(ctx, id.ObjectId)
	if err != nil {
		if status == http.StatusNotFound {
			return tf.ErrorDiagPathF(nil, "application_object_id", "Application with object ID %q was not found", id.ObjectId)
		}
		return tf.ErrorDiagPathF(err, "application_object_id", "Retrieving Application with object ID %q", id.ObjectId)
	}

	var credential *models.KeyCredential
	if app.KeyCredentials != nil {
		for _, cred := range *app.KeyCredentials {
			if cred.KeyId != nil && *cred.KeyId == id.KeyId {
				credential = &cred
				break
			}
		}
	}

	if credential == nil {
		log.Printf("[DEBUG] Certificate credential %q (ID %q) was not found - removing from state!", id.KeyId, id.ObjectId)
		d.SetId("")
		return nil
	}

	if dg := tf.Set(d, "application_object_id", id.ObjectId); dg != nil {
		return dg
	}

	if dg := tf.Set(d, "key_id", id.KeyId); dg != nil {
		return dg
	}

	keyType := ""
	if v := credential.Type; v != nil {
		keyType = *v
	}
	if dg := tf.Set(d, "type", keyType); dg != nil {
		return dg
	}

	startDate := ""
	if v := credential.StartDateTime; v != nil {
		startDate = v.Format(time.RFC3339)
	}
	if dg := tf.Set(d, "start_date", startDate); dg != nil {
		return dg
	}

	endDate := ""
	if v := credential.EndDateTime; v != nil {
		endDate = v.Format(time.RFC3339)
	}
	if dg := tf.Set(d, "end_date", endDate); dg != nil {
		return dg
	}

	return nil
}

func applicationCertificateResourceDeleteMsGraph(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client).Applications.MsClient

	id, err := parse.CertificateID(d.Id())
	if err != nil {
		return tf.ErrorDiagPathF(err, "id", "Parsing certificate credential with ID %q", d.Id())
	}

	tf.LockByName(resourceApplicationName, id.ObjectId)
	defer tf.UnlockByName(resourceApplicationName, id.ObjectId)

	app, status, err := client.Get(ctx, id.ObjectId)
	if err != nil {
		if status == http.StatusNotFound {
			log.Printf("[DEBUG] Application with Object ID %q was not found - removing from state!", id.ObjectId)
			return nil
		}
		return tf.ErrorDiagPathF(err, "application_object_id", "Retrieving Application with object ID %q", id.ObjectId)
	}

	newCredentials := make([]models.KeyCredential, 0)
	if app.KeyCredentials != nil {
		for _, cred := range *app.KeyCredentials {
			if cred.KeyId != nil && *cred.KeyId != id.KeyId {
				newCredentials = append(newCredentials, cred)
			}
		}
	}

	properties := models.Application{
		ID:             &id.ObjectId,
		KeyCredentials: &newCredentials,
	}
	if _, err := client.Update(ctx, properties); err != nil {
		return tf.ErrorDiagF(err, "Removing certificate credential %q from application with object ID %q", id.KeyId, id.ObjectId)
	}

	return nil
}
