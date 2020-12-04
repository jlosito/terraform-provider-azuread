package aadgraph_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/terraform-providers/terraform-provider-azuread/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azuread/internal/acceptance/check"
	"github.com/terraform-providers/terraform-provider-azuread/internal/clients"
	"github.com/terraform-providers/terraform-provider-azuread/internal/utils"
)

type GroupResource struct{}

func TestAccGroup_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_group", "test")
	r := GroupResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccGroup_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_group", "test")
	r := GroupResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.complete(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccGroup_owners(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_group", "test")
	r := GroupResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.withThreeOwners(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccGroup_members(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_group", "test")
	r := GroupResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.withThreeMembers(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccGroup_membersAndOwners(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_group", "test")
	r := GroupResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.withOwnersAndMembers(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccGroup_membersDiverse(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_group", "test")
	r := GroupResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.withDiverseMembers(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccGroup_ownersDiverse(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_group", "test")
	r := GroupResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.withDiverseOwners(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccGroup_membersUpdate(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_group", "test")
	r := GroupResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.withOneMember(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.withThreeMembers(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.withServicePrincipalMember(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.noMembers(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccGroup_ownersUpdate(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_group", "test")
	r := GroupResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.withThreeOwners(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.withOneOwner(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.withServicePrincipalOwner(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.noOwners(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccGroup_preventDuplicateNamesPass(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_group", "test")
	r := GroupResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.preventDuplicateNamesPass(data),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(data.ResourceName, "name", fmt.Sprintf("acctestGroup-%d", data.RandomInteger)),
			),
		},
		data.ImportStep("prevent_duplicate_names"),
	})
}

func TestAccGroup_preventDuplicateNamesFail(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_group", "test")
	r := GroupResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config:      r.preventDuplicateNamesFail(data),
			ExpectError: regexp.MustCompile("existing Group .+ was found"),
		},
	})
}

func (r GroupResource) Exists(ctx context.Context, clients *clients.AadClient, state *terraform.InstanceState) (*bool, error) {
	resp, err := clients.AadGraph.GroupsClient.Get(ctx, state.ID)

	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return nil, fmt.Errorf("Group with object ID %q does not exist", state.ID)
		}
		return nil, fmt.Errorf("failed to retrieve Group with object ID %q: %+v", state.ID, err)
	}

	return utils.Bool(resp.ObjectID != nil && *resp.ObjectID == state.ID), nil
}

func (GroupResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
resource "azuread_group" "test" {
  name = "acctestGroup-%[1]d"
}
`, data.RandomInteger)
}

func (GroupResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azuread_group" "test" {
  name        = "acctestGroup-%[2]d"
  description = "Please delete me as this is a.test.AD group!"
  members     = [azuread_user.test.object_id]
  owners      = [azuread_user.test.object_id]
}
`, UserResource{}.basic(data), data.RandomInteger)
}

func (GroupResource) noMembers(data acceptance.TestData) string {
	return fmt.Sprintf(`
resource "azuread_group" "test" {
  name    = "acctestGroup-%[1]d"
  members = []
}
`, data.RandomInteger)
}

func (GroupResource) noOwners(data acceptance.TestData) string {
	return fmt.Sprintf(`
resource "azuread_group" "test" {
  name   = "acctestGroup-%[1]d"
  owners = []
}
`, data.RandomInteger)
}

func (GroupResource) diverseDirectoryObjects(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s
%[2]s

resource "azuread_group" "member" {
  name = "acctestGroup-%[3]d-Member"
}
`, ServicePrincipalResource{}.basic(data), UserResource{}.basic(data), data.RandomInteger)
}

func (r GroupResource) withDiverseMembers(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azuread_group" "test" {
  name    = "acctestGroup-%[2]d"
  members = [azuread_user.test.object_id, azuread_group.member.object_id, azuread_service_principal.test.object_id]
}
`, r.diverseDirectoryObjects(data), data.RandomInteger)
}

func (r GroupResource) withDiverseOwners(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azuread_group" "test" {
  name   = "acctestGroup-%[2]d"
  owners = [azuread_user.test.object_id, azuread_service_principal.test.object_id]
}
`, r.diverseDirectoryObjects(data), data.RandomInteger)
}

func (GroupResource) withOneMember(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azuread_group" "test" {
  name    = "acctestGroup-%[2]d"
  members = [azuread_user.testA.object_id]
}
`, UserResource{}.threeUsersABC(data), data.RandomInteger)
}

func (GroupResource) withOneOwner(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azuread_group" "test" {
  name   = "acctestGroup-%[2]d"
  owners = [azuread_user.testA.object_id]
}
`, UserResource{}.threeUsersABC(data), data.RandomInteger)
}

func (GroupResource) withThreeMembers(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azuread_group" "test" {
  name    = "acctestGroup-%[2]d"
  members = [azuread_user.testA.object_id, azuread_user.testB.object_id, azuread_user.testC.object_id]
}
`, UserResource{}.threeUsersABC(data), data.RandomInteger)
}

func (GroupResource) withThreeOwners(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azuread_group" "test" {
  name   = "acctestGroup-%[2]d"
  owners = [azuread_user.testA.object_id, azuread_user.testB.object_id, azuread_user.testC.object_id]
}
`, UserResource{}.threeUsersABC(data), data.RandomInteger)
}

func (GroupResource) withOwnersAndMembers(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azuread_group" "test" {
  name    = "acctestGroup-%[2]d"
  owners  = [azuread_user.testA.object_id]
  members = [azuread_user.testB.object_id, azuread_user.testC.object_id]
}
`, UserResource{}.threeUsersABC(data), data.RandomInteger)
}

func (GroupResource) withServicePrincipalMember(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azuread_group" "test" {
  name    = "acctestGroup-%[2]d"
  members = [azuread_service_principal.test.object_id]
}
`, ServicePrincipalResource{}.basic(data), data.RandomInteger)
}

func (GroupResource) withServicePrincipalOwner(data acceptance.TestData) string {
	return fmt.Sprintf(`
resource "azuread_application" "test" {
  name = "acctestApp-%[1]d"
}

resource "azuread_service_principal" "test" {
  application_id = azuread_application.test.application_id
}

resource "azuread_group" "test" {
  name   = "acctestGroup-%[1]d"
  owners = [azuread_service_principal.test.object_id]
}
`, data.RandomInteger)
}

func (GroupResource) preventDuplicateNamesPass(data acceptance.TestData) string {
	return fmt.Sprintf(`
resource "azuread_group" "test" {
  name                    = "acctestGroup-%[1]d"
  prevent_duplicate_names = true
}
`, data.RandomInteger)
}

func (r GroupResource) preventDuplicateNamesFail(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azuread_group" "duplicate" {
  name                    = azuread_group.test.name
  prevent_duplicate_names = true
}
`, r.basic(data))
}
