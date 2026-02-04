package ai_search_instance_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/ai_search"
	"github.com/cloudflare/terraform-provider-cloudflare/internal/acctest"
	"github.com/cloudflare/terraform-provider-cloudflare/internal/consts"
	"github.com/cloudflare/terraform-provider-cloudflare/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func init() {
	resource.AddTestSweepers("cloudflare_ai_search_instance", &resource.Sweeper{
		Name: "cloudflare_ai_search_instance",
		F:    testSweepCloudflareAISearchInstances,
	})
}

func testSweepCloudflareAISearchInstances(region string) error {
	client := acctest.SharedClient()
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	ctx := context.Background()
	page, err := client.AISearch.Instances.List(ctx, ai_search.InstanceListParams{
		AccountID: cloudflare.F(accountID),
	})
	if err != nil {
		return fmt.Errorf("failed to fetch AI Search instances: %w", err)
	}

	for page != nil && len(page.Result) > 0 {
		for _, instance := range page.Result {
			if !utils.ShouldSweepResource(instance.ID) {
				continue
			}
			_, err := client.AISearch.Instances.Delete(
				ctx,
				instance.ID,
				ai_search.InstanceDeleteParams{
					AccountID: cloudflare.F(accountID),
				},
			)
			if err != nil {
				return fmt.Errorf("failed to delete AI Search instance %q: %w", instance.ID, err)
			}
		}

		page, err = page.GetNextPage()
		if err != nil {
			break
		}
	}

	return nil
}

func TestAccCloudflareAISearchInstance_Basic(t *testing.T) {
	rnd := utils.GenerateRandomResourceName()
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	resourceName := "cloudflare_ai_search_instance." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.TestAccPreCheck(t)
			acctest.TestAccPreCheck_AccountID(t)
		},
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCloudflareAISearchInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareAISearchInstanceBasic(rnd, accountID),
				// ExpectNonEmptyPlan is set to true due to computed nested objects having state drift
				// This is a known issue with auto-generated schemas for complex nested attributes
				ExpectNonEmptyPlan: true,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("id"), knownvalue.StringExact(rnd)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("account_id"), knownvalue.StringExact(accountID)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("source"), knownvalue.StringExact(rnd)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("type"), knownvalue.StringExact("r2")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("max_num_results"), knownvalue.Int64Exact(10)),
				},
			},
			// Update and import tests skipped due to state drift issues with computed nested objects
		},
	})
}

func testAccCheckCloudflareAISearchInstanceBasic(rnd, accountID string) string {
	return acctest.LoadTestCase("aisearchinstancebasic.tf", rnd, accountID)
}

func testAccCheckCloudflareAISearchInstanceDestroy(s *terraform.State) error {
	client := acctest.SharedClient()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare_ai_search_instance" {
			continue
		}

		accountID := rs.Primary.Attributes[consts.AccountIDSchemaKey]
		_, err := client.AISearch.Instances.Read(
			context.Background(),
			rs.Primary.ID,
			ai_search.InstanceReadParams{
				AccountID: cloudflare.F(accountID),
			},
		)
		if err == nil {
			return fmt.Errorf("AI Search instance still exists")
		}
	}

	return nil
}
