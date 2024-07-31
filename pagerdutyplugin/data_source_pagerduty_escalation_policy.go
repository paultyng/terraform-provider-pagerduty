package pagerduty

import (
	"context"
	"fmt"
	"log"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/PagerDuty/terraform-provider-pagerduty/util/apiutil"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSourceEscalationPolicy struct{ client *pagerduty.Client }

var _ datasource.DataSourceWithConfigure = (*dataSourceEscalationPolicy)(nil)

func (*dataSourceEscalationPolicy) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "pagerduty_escalation_policy"
}

func (*dataSourceEscalationPolicy) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":   schema.StringAttribute{Computed: true},
			"name": schema.StringAttribute{Required: true},
		},
	}
}

func (d *dataSourceEscalationPolicy) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.Append(ConfigurePagerdutyClient(&d.client, req.ProviderData)...)
}

func (d *dataSourceEscalationPolicy) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	log.Println("[INFO] Reading PagerDuty escalation policy")

	var searchName types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("name"), &searchName)...)
	if resp.Diagnostics.HasError() {
		return
	}
	opts := pagerduty.ListEscalationPoliciesOptions{Query: searchName.ValueString()}

	var found *pagerduty.EscalationPolicy
	err := apiutil.All(ctx, func(offset int) (bool, error) {
		resp, err := d.client.ListEscalationPoliciesWithContext(ctx, opts)
		if err != nil {
			return false, err
		}

		for _, escalationPolicy := range resp.EscalationPolicies {
			if escalationPolicy.Name == searchName.ValueString() {
				found = &escalationPolicy
				return false, nil
			}
		}

		return resp.More, nil
	})
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error searching PagerDuty escalation policy %s", searchName),
			err.Error(),
		)
		return
	}

	if found == nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to locate any escalation policy with the name: %s", searchName),
			"",
		)
		return
	}

	model := dataSourceEscalationPolicyModel{
		ID:   types.StringValue(found.ID),
		Name: types.StringValue(found.Name),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

type dataSourceEscalationPolicyModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}
