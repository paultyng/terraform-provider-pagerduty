package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/terraform-providers/terraform-provider-pagerduty/pagerduty"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: pagerduty.Provider})
}
