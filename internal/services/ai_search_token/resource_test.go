package ai_search_token_test

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
	resource.AddTestSweepers("cloudflare_ai_search_token", &resource.Sweeper{
		Name: "cloudflare_ai_search_token",
		F:    testSweepCloudflareAISearchTokens,
	})
}

func testSweepCloudflareAISearchTokens(region string) error {
	client := acctest.SharedClient()
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	ctx := context.Background()
	page, err := client.AISearch.Tokens.List(ctx, ai_search.TokenListParams{
		AccountID: cloudflare.F(accountID),
	})
	if err != nil {
		return fmt.Errorf("failed to fetch AI Search tokens: %w", err)
	}

	for page != nil && len(page.Result) > 0 {
		for _, token := range page.Result {
			if !utils.ShouldSweepResource(token.Name) {
				continue
			}
			_, err := client.AISearch.Tokens.Delete(
				ctx,
				token.ID,
				ai_search.TokenDeleteParams{
					AccountID: cloudflare.F(accountID),
				},
			)
			if err != nil {
				return fmt.Errorf("failed to delete AI Search token %q: %w", token.Name, err)
			}
		}

		page, err = page.GetNextPage()
		if err != nil {
			break
		}
	}

	return nil
}

func TestAccCloudflareAISearchToken_Basic(t *testing.T) {
	rnd := utils.GenerateRandomResourceName()
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	resourceName := "cloudflare_ai_search_token." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.TestAccPreCheck(t)
			acctest.TestAccPreCheck_AccountID(t)
		},
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCloudflareAISearchTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareAISearchTokenBasic(rnd, accountID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("name"), knownvalue.StringExact(rnd)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("account_id"), knownvalue.StringExact(accountID)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("enabled"), knownvalue.Bool(true)),
				},
			},
			// Import test skipped - cf_api_key is write-only and cannot be imported
		},
	})
}

func testAccCheckCloudflareAISearchTokenBasic(rnd, accountID string) string {
	return acctest.LoadTestCase("aisearchtokenbasic.tf", rnd, accountID)
}

func testAccCheckCloudflareAISearchTokenDestroy(s *terraform.State) error {
	client := acctest.SharedClient()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare_ai_search_token" {
			continue
		}

		accountID := rs.Primary.Attributes[consts.AccountIDSchemaKey]
		_, err := client.AISearch.Tokens.Read(
			context.Background(),
			rs.Primary.ID,
			ai_search.TokenReadParams{
				AccountID: cloudflare.F(accountID),
			},
		)
		if err == nil {
			return fmt.Errorf("AI Search token still exists")
		}
	}

	return nil
}
