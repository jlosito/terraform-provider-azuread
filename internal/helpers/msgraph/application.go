package msgraph

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/manicminer/hamilton/clients"
	"github.com/manicminer/hamilton/models"

	"github.com/terraform-providers/terraform-provider-azuread/internal/utils"
)

func ApplicationFindByName(ctx context.Context, client *clients.ApplicationsClient, displayName string) (*models.Application, error) {
	filter := fmt.Sprintf("displayName eq '%s'", displayName)
	result, _, err := client.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("unable to list Applications with filter %q: %+v", filter, err)
	}

	if result != nil {
		for _, app := range *result {
			if app.DisplayName != nil && *app.DisplayName == displayName {
				return &app, nil
			}
		}
	}

	return nil, nil
}

func ApplicationSetAppRoles(ctx context.Context, client *clients.ApplicationsClient, application *models.Application, newRoles *[]models.ApplicationAppRole) error {
	if application.ID == nil {
		return fmt.Errorf("Cannot use Application model with nil ID")
	}

	if newRoles == nil {
		newRoles = &[]models.ApplicationAppRole{}
	}

	// Roles must be disabled before they can be edited or removed.
	// Since we cannot match them by ID, we have to disable all the roles, and replace them in one pass.
	app, status, err := client.Get(ctx, *application.ID)
	if err != nil {
		if status == http.StatusNotFound {
			return fmt.Errorf("application with ID %q was not found", *application.ID)
		}

		return fmt.Errorf("retrieving Application with object ID %q: %+v", *application.ID, err)
	}

	// don't update if no changes to be made
	if app.AppRoles != nil && reflect.DeepEqual(*app.AppRoles, *newRoles) {
		return nil
	}

	// first disable any existing roles
	if app.AppRoles != nil && len(*app.AppRoles) > 0 {
		properties := models.Application{
			ID:       application.ID,
			AppRoles: app.AppRoles,
		}

		for _, role := range *properties.AppRoles {
			*role.IsEnabled = false
		}

		if _, err := client.Update(ctx, properties); err != nil {
			return fmt.Errorf("disabling App Roles for Application with object ID %q: %+v", *application.ID, err)
		}
	}

	// then set the new roles
	properties := models.Application{
		ID:       application.ID,
		AppRoles: newRoles,
	}

	if _, err := client.Update(ctx, properties); err != nil {
		return fmt.Errorf("setting App Roles for Application with object ID %q: %+v", *application.ID, err)
	}

	return nil
}

func ApplicationSetOAuth2PermissionScopes(ctx context.Context, client *clients.ApplicationsClient, application *models.Application, newScopes *[]models.ApplicationApiPermissionScope) error {
	if application.ID == nil {
		return fmt.Errorf("Cannot use Application model with nil ID")
	}

	if newScopes == nil {
		newScopes = &[]models.ApplicationApiPermissionScope{}
	}

	// OAuth2 Permission Scopes must be disabled before they can be edited or removed.
	// Since we cannot match them by ID, we have to disable all the scopes, and replace them in one pass.
	app, status, err := client.Get(ctx, *application.ID)
	if err != nil {
		if status == http.StatusNotFound {
			return fmt.Errorf("application with ID %q was not found", *application.ID)
		}

		return fmt.Errorf("retrieving Application with object ID %q: %+v", *application.ID, err)
	}

	// don't update if no changes to be made
	if app.Api != nil && app.Api.OAuth2PermissionScopes != nil && reflect.DeepEqual(*app.Api.OAuth2PermissionScopes, *newScopes) {
		return nil
	}

	// first disable any existing scopes
	if app.Api != nil && app.Api.OAuth2PermissionScopes != nil && len(*app.Api.OAuth2PermissionScopes) > 0 {
		properties := models.Application{
			ID: application.ID,
			Api: &models.ApplicationApi{
				OAuth2PermissionScopes: app.Api.OAuth2PermissionScopes,
			},
		}

		for _, scope := range *properties.Api.OAuth2PermissionScopes {
			*scope.IsEnabled = false
		}

		if _, err := client.Update(ctx, properties); err != nil {
			return fmt.Errorf("disabling OAuth2 Permission Scopes for Application with object ID %q: %+v", *application.ID, err)
		}
	}

	// then set the new scopes
	properties := models.Application{
		ID: application.ID,
		Api: &models.ApplicationApi{
			OAuth2PermissionScopes: newScopes,
		},
	}

	if _, err := client.Update(ctx, properties); err != nil {
		return fmt.Errorf("setting OAuth2 Permission Scopes for Application with object ID %q: %+v", *application.ID, err)
	}

	return nil
}

func ApplicationSetOwners(ctx context.Context, client *clients.ApplicationsClient, application *models.Application, desiredOwners []string) error {
	if application.ID == nil {
		return fmt.Errorf("Cannot use Application model with nil ID")
	}

	owners, _, err := client.ListOwners(ctx, *application.ID)
	if err != nil {
		return fmt.Errorf("retrieving owners for Application with object ID %q: %+v", *application.ID, err)
	}

	existingOwners := *owners
	ownersForRemoval := utils.Difference(existingOwners, desiredOwners)
	ownersToAdd := utils.Difference(desiredOwners, existingOwners)

	if ownersForRemoval != nil {
		if _, err = client.RemoveOwners(ctx, *application.ID, &ownersForRemoval); err != nil {
			return fmt.Errorf("removing owner from Application with object ID %q: %+v", *application.ID, err)
		}
	}

	if ownersToAdd != nil {
		for _, m := range ownersToAdd {
			application.AppendOwner(client.BaseClient.Endpoint, client.BaseClient.ApiVersion, m)
		}

		if _, err := client.AddOwners(ctx, application); err != nil {
			return err
		}
	}
	return nil
}

func AppRoleFindById(app *models.Application, roleId string) (*models.ApplicationAppRole, error) {
	if app == nil || app.AppRoles == nil {
		return nil, nil
	}

	if roleId == "" {
		return nil, errors.New("specified role ID is blank")
	}

	for _, r := range *app.AppRoles {
		if r.ID == nil {
			continue
		}
		if *r.ID == roleId {
			return &r, nil
		}
	}

	return nil, nil
}
