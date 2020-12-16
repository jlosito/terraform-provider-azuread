package applications

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/manicminer/hamilton/models"

	"github.com/terraform-providers/terraform-provider-azuread/internal/clients"
	"github.com/terraform-providers/terraform-provider-azuread/internal/tf"
)

func applicationDataSourceReadMsGraph(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client).Applications.MsClient

	var app *models.Application

	if objectId, ok := d.Get("object_id").(string); ok && objectId != "" {
		var status int
		var err error
		app, status, err = client.Get(ctx, objectId)
		if err != nil {
			if status == http.StatusNotFound {
				return tf.ErrorDiagPathF(nil, "object_id", "Application with object ID %q was not found", objectId)
			}

			return tf.ErrorDiagPathF(err, "application_object_id", "Retrieving Application with object ID %q", objectId)
		}
	} else {
		var fieldName, fieldValue string
		if applicationId, ok := d.Get("application_id").(string); ok && applicationId != "" {
			fieldName = "appId"
			fieldValue = applicationId
		} else if displayName, ok := d.Get("display_name").(string); ok && displayName != "" {
			fieldName = "displayName"
			fieldValue = displayName
		} else if name, ok := d.Get("name").(string); ok && name != "" {
			fieldName = "displayName"
			fieldValue = name
		} else {
			return tf.ErrorDiagF(nil, "One of `object_id`, `application_id` or `displayName` must be specified")
		}

		filter := fmt.Sprintf("%s eq '%s'", fieldName, fieldValue)

		result, _, err := client.List(ctx, filter)
		if err != nil {
			return tf.ErrorDiagF(err, "Listing applications for filter %q", filter)
		}

		switch {
		case result == nil || len(*result) == 0:
			return tf.ErrorDiagF(fmt.Errorf("No applications found matching filter: %q", filter), "Application not found")
		case len(*result) > 1:
			return tf.ErrorDiagF(fmt.Errorf("Found multiple applications matching filter: %q", filter), "Multiple applications found")
		}

		app = &(*result)[0]
		switch fieldName {
		case "appId":
			if app.AppId == nil {
				return tf.ErrorDiagF(fmt.Errorf("nil AppID for applications matching filter: %q", filter), "Bad API Response")
			}
			if *app.AppId != fieldValue {
				return tf.ErrorDiagF(fmt.Errorf("AppID does not match (%q != %q) for applications matching filter: %q", *app.AppId, fieldValue, filter), "Bad API Response")
			}
		case "displayName":
			if app.DisplayName == nil {
				return tf.ErrorDiagF(fmt.Errorf("nil displayName for applications matching filter: %q", filter), "Bad API Response")
			}
			if *app.DisplayName != fieldValue {
				return tf.ErrorDiagF(fmt.Errorf("DisplayName does not match (%q != %q) for applications matching filter: %q", *app.DisplayName, fieldValue, filter), "Bad API Response")
			}
		}
	}

	if app.ID == nil {
		return tf.ErrorDiagF(fmt.Errorf("Object ID returned for application is nil"), "Bad API Response")
	}

	d.SetId(*app.ID)

	if dg := tf.Set(d, "object_id", app.ID); dg != nil {
		return dg
	}

	if dg := tf.Set(d, "application_id", app.AppId); dg != nil {
		return dg
	}

	if dg := tf.Set(d, "app_roles", flattenApplicationAppRoles(app.AppRoles)); dg != nil {
		return dg
	}

	availableToOtherTenants := app.SignInAudience == models.SignInAudienceAzureADMultipleOrgs
	if dg := tf.Set(d, "available_to_other_tenants", availableToOtherTenants); dg != nil {
		return dg
	}

	if dg := tf.Set(d, "display_name", app.DisplayName); dg != nil {
		return dg
	}

	if dg := tf.Set(d, "name", app.DisplayName); dg != nil {
		return dg
	}

	if dg := tf.Set(d, "group_membership_claims", app.GroupMembershipClaims); dg != nil {
		return dg
	}

	if dg := tf.Set(d, "identifier_uris", tf.FlattenStringSlicePtr(app.IdentifierUris)); dg != nil {
		return dg
	}

	if dg := tf.Set(d, "optional_claims", flattenApplicationOptionalClaims(app.OptionalClaims)); dg != nil {
		return dg
	}

	if dg := tf.Set(d, "required_resource_access", flattenApplicationRequiredResourceAccess(app.RequiredResourceAccess)); dg != nil {
		return dg
	}

	var appType string
	if v := app.IsFallbackPublicClient; v != nil && *v {
		appType = "native"
	} else {
		appType = "webapp/api"
	}

	if dg := tf.Set(d, "type", appType); dg != nil {
		return dg
	}

	if app.Api != nil {
		if dg := tf.Set(d, "oauth2_permissions", flattenApplicationOAuth2Permissions(app.Api.OAuth2PermissionScopes)); dg != nil {
			return dg
		}
	}

	if app.Web != nil {
		if dg := tf.Set(d, "homepage", app.Web.HomePageUrl); dg != nil {
			return dg
		}

		if dg := tf.Set(d, "logout_url", app.Web.LogoutUrl); dg != nil {
			return dg
		}

		if dg := tf.Set(d, "reply_urls", tf.FlattenStringSlicePtr(app.Web.RedirectUris)); dg != nil {
			return dg
		}

		if app.Web.ImplicitGrantSettings != nil {
			if dg := tf.Set(d, "oauth2_allow_implicit_flow", app.Web.ImplicitGrantSettings.EnableAccessTokenIssuance); dg != nil {
				return dg
			}
		}
	}

	owners, _, err := client.ListOwners(ctx, *app.ID)
	if err != nil {
		return tf.ErrorDiagPathF(err, "owners", "Could not retrieve owners for application with object ID %q", *app.ID)
	}

	if dg := tf.Set(d, "owners", owners); dg != nil {
		return dg
	}

	return nil
}
