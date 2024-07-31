package pagerduty

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/PagerDuty/terraform-provider-pagerduty/util"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

type dataSourceUsers struct{ client *pagerduty.Client }

var _ datasource.DataSourceWithConfigure = (*dataSourceUsers)(nil)

func (*dataSourceUsers) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "pagerduty_users"
}

func (*dataSourceUsers) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"team_ids": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
			"users": schema.ListAttribute{
				Computed:    true,
				Description: "List of users who are members of the team",
				ElementType: userObjectType,
			},
		},
	}
}

func (d *dataSourceUsers) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.Append(ConfigurePagerdutyClient(&d.client, req.ProviderData)...)
}

func (d *dataSourceUsers) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	log.Println("[INFO] Reading PagerDuty users")

	var list types.List
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("team_ids"), &list)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var teamIdStrings []types.String
	resp.Diagnostics.Append(list.ElementsAs(ctx, &teamIdStrings, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamIds := make([]string, 0, len(list.Elements()))
	for _, v := range teamIdStrings {
		teamIds = append(teamIds, v.ValueString())
	}
	opts := pagerduty.ListUsersOptions{TeamIDs: teamIds}

	var model dataSourceUsersModel
	err := retry.RetryContext(ctx, 2*time.Minute, func() *retry.RetryError {
		response, err := d.client.ListUsersWithContext(ctx, opts)
		if err != nil {
			if util.IsBadRequestError(err) {
				return retry.NonRetryableError(err)
			}
			return retry.RetryableError(err)
		}
		model = flattenUsers(response.Users, list)
		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading PagerDuty users", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

type dataSourceUsersModel struct {
	ID      types.String `tfsdk:"id"`
	Users   types.List   `tfsdk:"users"`
	TeamIDs types.List   `tfsdk:"team_ids"`
}

var userObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"id":          types.StringType,
		"description": types.StringType,
		"email":       types.StringType,
		"job_title":   types.StringType,
		"name":        types.StringType,
		"role":        types.StringType,
		"time_zone":   types.StringType,
		"type":        types.StringType,
	},
}

func flattenUsers(list []pagerduty.User, teamIds types.List) dataSourceUsersModel {
	userValues := make([]attr.Value, 0, len(list))
	for _, u := range list {
		obj := types.ObjectValueMust(userObjectType.AttrTypes, map[string]attr.Value{
			"id":          types.StringValue(u.ID),
			"name":        types.StringValue(u.Name),
			"email":       types.StringValue(u.Email),
			"role":        types.StringValue(u.Role),
			"job_title":   types.StringValue(u.JobTitle),
			"time_zone":   types.StringValue(u.Timezone),
			"description": types.StringValue(u.Description),
			"type":        types.StringNull(),
		})
		userValues = append(userValues, obj)
	}
	return dataSourceUsersModel{
		ID:      types.StringValue(strconv.FormatInt(time.Now().Unix(), 10)),
		Users:   types.ListValueMust(userObjectType, userValues),
		TeamIDs: teamIds,
	}
}
