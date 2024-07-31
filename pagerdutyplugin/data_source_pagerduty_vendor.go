package pagerduty

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/PagerDuty/terraform-provider-pagerduty/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

type dataSourceVendor struct{ client *pagerduty.Client }

var _ datasource.DataSourceWithConfigure = (*dataSourceVendor)(nil)

func (*dataSourceVendor) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "pagerduty_vendor"
}

func (*dataSourceVendor) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":   schema.StringAttribute{Computed: true},
			"name": schema.StringAttribute{Required: true},
			"type": schema.StringAttribute{Computed: true},
		},
	}
}

func (d *dataSourceVendor) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.Append(ConfigurePagerdutyClient(&d.client, req.ProviderData)...)
}

func (d *dataSourceVendor) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	log.Println("[INFO] Reading PagerDuty vendor")

	var searchName types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("name"), &searchName)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var found *pagerduty.Vendor
	err := retry.RetryContext(ctx, 2*time.Minute, func() *retry.RetryError {
		list, err := d.client.ListVendorsWithContext(ctx, pagerduty.ListVendorOptions{})
		if err != nil {
			if util.IsBadRequestError(err) {
				return retry.NonRetryableError(err)
			}
			return retry.RetryableError(err)
		}

		for _, vendor := range list.Vendors {
			if strings.EqualFold(vendor.Name, searchName.ValueString()) {
				found = &vendor
				break
			}
		}

		// We didn't find an exact match, so let's fallback to partial matching.
		if found == nil {
			pr := regexp.MustCompile("(?i)" + searchName.ValueString())
			for _, vendor := range list.Vendors {
				if pr.MatchString(vendor.Name) {
					found = &vendor
					break
				}
			}
		}
		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading PagerDuty vendor %s", searchName),
			err.Error(),
		)
		return
	}

	if found == nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to locate any vendor with the name: %s", searchName),
			"",
		)
		return
	}

	model := dataSourceVendorModel{
		ID:   types.StringValue(found.ID),
		Name: types.StringValue(found.Name),
		Type: types.StringValue(found.GenericServiceType),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

type dataSourceVendorModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
}
