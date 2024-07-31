package pagerduty

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/PagerDuty/terraform-provider-pagerduty/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

type dataSourceUser struct{ client *pagerduty.Client }

var _ datasource.DataSourceWithConfigure = (*dataSourceUser)(nil)

func (*dataSourceUser) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "pagerduty_user"
}

func (*dataSourceUser) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"email":       schema.StringAttribute{Required: true},
			"description": schema.StringAttribute{Computed: true},
			"id":          schema.StringAttribute{Computed: true},
			"job_title":   schema.StringAttribute{Computed: true},
			"name":        schema.StringAttribute{Computed: true},
			"role":        schema.StringAttribute{Computed: true},
			"time_zone":   schema.StringAttribute{Computed: true},
		},
	}
}

func (d *dataSourceUser) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.Append(ConfigurePagerdutyClient(&d.client, req.ProviderData)...)
}

func (d *dataSourceUser) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	log.Println("[INFO] Reading PagerDuty user")

	var searchEmail types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("email"), &searchEmail)...)
	if resp.Diagnostics.HasError() {
		return
	}
	opts := pagerduty.ListUsersOptions{Query: searchEmail.ValueString()}

	var found *pagerduty.User
	err := retry.RetryContext(ctx, 2*time.Minute, func() *retry.RetryError {
		response, err := d.client.ListUsersWithContext(ctx, opts)
		if err != nil {
			if util.IsBadRequestError(err) {
				return retry.NonRetryableError(err)
			}
			return retry.RetryableError(err)
		}

		for _, user := range response.Users {
			// if strings.EqualFold(user.Email, searchEmail.ValueString()) {
			if user.Email == searchEmail.ValueString() {
				found = &user
				break
			}
		}
		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading PagerDuty user %s", searchEmail),
			err.Error(),
		)
		return
	}

	if found == nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to locate any user with the email: %s", searchEmail),
			"",
		)
		return
	}

	model := dataSourceUserModel{
		Email:       types.StringValue(found.Email),
		Description: types.StringValue(found.Description),
		ID:          types.StringValue(found.ID),
		JobTitle:    types.StringValue(found.JobTitle),
		Name:        types.StringValue(found.Name),
		Role:        types.StringValue(found.Role),
		Timezone:    types.StringValue(found.Timezone),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

type dataSourceUserModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Email       types.String `tfsdk:"email"`
	Description types.String `tfsdk:"description"`
	JobTitle    types.String `tfsdk:"job_title"`
	Role        types.String `tfsdk:"role"`
	Timezone    types.String `tfsdk:"time_zone"`
}
