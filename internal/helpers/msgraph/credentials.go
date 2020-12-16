package msgraph

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/manicminer/hamilton/models"

	"github.com/terraform-providers/terraform-provider-azuread/internal/utils"
)

func KeyCredentialForResource(d *schema.ResourceData) (*models.KeyCredential, error) {
	keyType := d.Get("type").(string)
	value := d.Get("value").(string)
	encodedValue := base64.StdEncoding.EncodeToString([]byte(value))

	var keyId string
	if v, ok := d.GetOk("key_id"); ok {
		keyId = v.(string)
	} else {
		kid, err := uuid.GenerateUUID()
		if err != nil {
			return nil, err
		}

		keyId = kid
	}

	var endDate time.Time
	if v := d.Get("end_date").(string); v != "" {
		var err error
		endDate, err = time.Parse(time.RFC3339, v)
		if err != nil {
			return nil, CredentialError{str: fmt.Sprintf("Unable to parse the provided end date %q: %+v", v, err), attr: "end_date"}
		}
	} else if v := d.Get("end_date_relative").(string); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, CredentialError{str: fmt.Sprintf("Unable to parse `end_date_relative` (%q) as a duration", v), attr: "end_date_relative"}
		}
		endDate = time.Now().Add(d)
	} else {
		return nil, CredentialError{str: "One of `end_date` or `end_date_relative` must be specified", attr: "end_date"}
	}

	credential := models.KeyCredential{
		KeyId:       utils.String(keyId),
		Type:        utils.String(keyType),
		Usage:       utils.String("verify"),
		Key:         utils.String(encodedValue),
		EndDateTime: &endDate,
	}

	if v, ok := d.GetOk("start_date"); ok {
		startDate, err := time.Parse(time.RFC3339, v.(string))
		if err != nil {
			return nil, CredentialError{str: fmt.Sprintf("Unable to parse the provided start date %q: %+v", v, err), attr: "start_date"}
		}
		credential.StartDateTime = &startDate
	}

	return &credential, nil
}
