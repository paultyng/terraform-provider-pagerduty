package pagerduty

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

// sharedConfigForRegion returns a common config setup needed for the sweeper
// functions for a given region
func sharedConfigForRegion(region string) (*Config, error) {
	if os.Getenv("PAGERDUTY_TOKEN") == "" {
		return nil, fmt.Errorf("$PAGERDUTY_TOKEN must be set")
	}

	config := &Config{
		Token: os.Getenv("PAGERDUTY_TOKEN"),
	}

	return config, nil
}
