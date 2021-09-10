package pagerduty

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"pagerduty": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProviderImpl(t *testing.T) {
	var _ *schema.Provider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("PAGERDUTY_PARALLEL"); v != "" {
		t.Parallel()
	}

	if v := os.Getenv("PAGERDUTY_TOKEN"); v == "" {
		t.Fatal("PAGERDUTY_TOKEN must be set for acceptance tests")
	}
}

// timeNowInLoc returns the current time in the given location.
// If an error occurs when trying to load the location, we just return the current local time.
func timeNowInLoc(name string) time.Time {
	loc, err := time.LoadLocation(name)
	if err != nil {
		log.Printf("[WARN] Failed to load location: %s", err)
		return time.Now()
	}

	return time.Now().In(loc)
}

// timeNowInAccLoc returns the current time in the given location.
// The location defaults to Europe/Dublin but can be controlled by the PAGERDUTY_TIME_ZONE environment variable.
// The location must match the PagerDuty account time zone or diff issues might bubble up in tests.
func timeNowInAccLoc() time.Time {
	name := "Europe/Dublin"

	if v := os.Getenv("PAGERDUTY_TIME_ZONE"); v != "" {
		name = v
	}

	return timeNowInLoc(name)
}

func testAccPreCheckPagerDutyAbility(t *testing.T, ability string) {
	if v := os.Getenv("PAGERDUTY_TOKEN"); v == "" {
		t.Fatal("PAGERDUTY_TOKEN must be set for acceptance tests")
	}

	config := &Config{
		Token: os.Getenv("PAGERDUTY_TOKEN"),
	}

	client, err := config.Client()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := client.Abilities.Test(ability); err != nil {
		t.Skipf("Missing ability: %s. Skipping test", ability)
	}
}
