package pagerduty

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/PagerDuty/terraform-provider-pagerduty/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

type resourceTeamMembership struct{ client *pagerduty.Client }

var (
	_ resource.ResourceWithConfigure   = (*resourceTeamMembership)(nil)
	_ resource.ResourceWithImportState = (*resourceTeamMembership)(nil)
)

func (r *resourceTeamMembership) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "pagerduty_team_membership"
}

func (r *resourceTeamMembership) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"team_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf("observer", "responder", "manager"),
				},
				Default: stringdefault.StaticString("manager"),
			},
		},
	}
}

func (r *resourceTeamMembership) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	model := requestAddTeamMembership(ctx, r.client, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id := model.ID.ValueString()
	role := model.Role.ValueString()

	model, err := requestGetTeamMembership(ctx, r.client, id, &role, RetryNotFound, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error creating PagerDuty team membership %s", model.ID),
			err.Error(),
		)
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *resourceTeamMembership) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var id types.String

	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)
	if resp.Diagnostics.HasError() {
		return
	}
	log.Printf("[INFO] Reading PagerDuty team membership %s", id)

	state, err := requestGetTeamMembership(ctx, r.client, id.ValueString(), nil, !RetryNotFound, &resp.Diagnostics)
	if err != nil {
		if util.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading PagerDuty team membership %s", id),
			err.Error(),
		)
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceTeamMembership) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	model := requestAddTeamMembership(ctx, r.client, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id := model.ID.ValueString()
	role := model.Role.ValueString()

	model, err := requestGetTeamMembership(ctx, r.client, id, &role, RetryNotFound, &resp.Diagnostics)
	if err != nil {
		if util.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error updating PagerDuty team membership %s", model.ID),
			err.Error(),
		)
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *resourceTeamMembership) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var id types.String

	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)
	if resp.Diagnostics.HasError() {
		return
	}
	log.Printf("[INFO] Deleting PagerDuty team membership %s", id)

	userID, teamID, err := util.ResourcePagerDutyParseColonCompoundID(id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Invalid Team Membership ID %s", id), err.Error())
		return
	}

	userIsInEP := false
	err = retry.RetryContext(ctx, 2*time.Minute, func() *retry.RetryError {
		if err := r.client.RemoveUserFromTeamWithContext(ctx, teamID, userID); err != nil {
			userIsInEP = strings.Contains(err.Error(), "User cannot be removed as they belong to an escalation policy on this team")
			if util.IsBadRequestError(err) {
				return retry.NonRetryableError(err)
			}
			return retry.RetryableError(err)
		}
		return nil
	})
	if userIsInEP {
		eps := fetchEscalationPoliciesWithUser(ctx, r.client, userID, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		diagnoseEscalationPoliciesAssociatedToUser(userID, teamID, eps, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if err != nil && !util.IsNotFoundError(err) {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error deleting PagerDuty team membership %s", id),
			err.Error(),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *resourceTeamMembership) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(ConfigurePagerdutyClient(&r.client, req.ProviderData)...)
}

func (r *resourceTeamMembership) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

type resourceTeamMembershipModel struct {
	ID     types.String `tfsdk:"id"`
	TeamID types.String `tfsdk:"team_id"`
	UserID types.String `tfsdk:"user_id"`
	Role   types.String `tfsdk:"role"`
}

func requestGetTeamMembership(ctx context.Context, client *pagerduty.Client, id string, neededRole *string, retryNotFound bool, diags *diag.Diagnostics) (resourceTeamMembershipModel, error) {
	var model resourceTeamMembershipModel

	userID, teamID, err := util.ResourcePagerDutyParseColonCompoundID(id)
	if err != nil {
		diags.AddError(fmt.Sprintf("Invalid Team Membership ID %s", id), err.Error())
		return model, nil
	}

	err = retry.RetryContext(ctx, 2*time.Minute, func() *retry.RetryError {
		resp, err := client.ListTeamMembers(ctx, teamID, pagerduty.ListTeamMembersOptions{Limit: 100})
		if err != nil {
			if util.IsBadRequestError(err) {
				return retry.NonRetryableError(err)
			}
			if !retryNotFound && util.IsNotFoundError(err) {
				return retry.NonRetryableError(err)
			}
			return retry.RetryableError(err)
		}

		for _, m := range resp.Members {
			if m.User.ID == userID {
				if neededRole != nil && m.Role != *neededRole {
					err = fmt.Errorf("Role %q fetched is different from configuration %q", m.Role, *neededRole)
					return retry.RetryableError(err)
				}
				model = flattenTeamMembership(userID, teamID, m.Role)
				return nil
			}
		}

		err = pagerduty.APIError{StatusCode: http.StatusNotFound}
		if retryNotFound {
			return retry.RetryableError(err)
		}
		return retry.NonRetryableError(err)
	})

	return model, err
}

func requestAddTeamMembership(ctx context.Context, client *pagerduty.Client, plan SchemaGetter, diags *diag.Diagnostics) resourceTeamMembershipModel {
	var model resourceTeamMembershipModel

	diags.Append(plan.Get(ctx, &model)...)
	if diags.HasError() {
		return model
	}
	opts, _ := buildPagerdutyTeamMembership(&model)
	log.Printf("[INFO] Creating PagerDuty team membership for user %s at team %s using role %s", opts.UserID, opts.TeamID, opts.Role)

	err := retry.RetryContext(ctx, 2*time.Minute, func() *retry.RetryError {
		err := client.AddUserToTeamWithContext(ctx, opts)
		if err != nil {
			if util.IsBadRequestError(err) {
				return retry.NonRetryableError(err)
			}
			return retry.RetryableError(err)
		}
		model.ID = flattenTeamMembershipID(opts.UserID, opts.TeamID)
		return nil
	})
	if err != nil {
		diags.AddError(
			fmt.Sprintf("Error creating PagerDuty team membership for user %s at team %s using role %s", opts.UserID, opts.TeamID, opts.Role),
			err.Error(),
		)
		return model
	}
	return model
}

func fetchEscalationPoliciesWithUser(ctx context.Context, client *pagerduty.Client, userID string, diags *diag.Diagnostics) []pagerduty.EscalationPolicy {
	var oncalls []pagerduty.OnCall
	err := retry.RetryContext(ctx, 2*time.Minute, func() *retry.RetryError {
		resp, err := client.ListOnCallsWithContext(ctx, pagerduty.ListOnCallOptions{UserIDs: []string{userID}, Limit: 100})
		if err != nil {
			if util.IsBadRequestError(err) {
				return retry.NonRetryableError(err)
			}
			return retry.RetryableError(err)
		}
		oncalls = resp.OnCalls
		return nil
	})
	if err != nil {
		diags.AddError(
			fmt.Sprintf("Error reading escalation policies for PagerDuty user %s", userID),
			err.Error(),
		)
		return nil
	}

	var eps []pagerduty.EscalationPolicy
	for _, oc := range oncalls {
		eps = append(eps, oc.EscalationPolicy)
	}

	return eps
}

func diagnoseEscalationPoliciesAssociatedToUser(userID, teamID string, eps []pagerduty.EscalationPolicy, diags *diag.Diagnostics) {
	if len(eps) == 0 {
		return // No diagnostics
	}

	pdURL, err := url.Parse(eps[0].HTMLURL)
	if err != nil {
		return // No diagnostics
	}

	var links []string
	for _, ep := range eps {
		links = append(links, fmt.Sprintf("\t* %s", ep.HTMLURL))
	}

	diags.AddError(
		fmt.Sprintf("User %q can't be removed from Team %q", userID, teamID),
		fmt.Sprintf(`As the user belongs to an Escalation Policy on this team. Please take one of the following remediation measures in order to unblock the Team Membership removal:
1. Remove the user from the following Escalation Policies:
%s
2. Remove the Escalation Policies from the Team:
	https://%s/teams/%s

After completing one of the above given remediation options come back to continue with the destruction of Team Membership.`,
			strings.Join(links, "\n"),
			pdURL.Hostname(),
			teamID,
		),
	)
}

func buildPagerdutyTeamMembership(model *resourceTeamMembershipModel) (pagerduty.AddUserToTeamOptions, string) {
	opts := pagerduty.AddUserToTeamOptions{
		TeamID: model.TeamID.ValueString(),
		UserID: model.UserID.ValueString(),
		Role:   pagerduty.TeamUserRole(model.Role.ValueString()),
	}
	return opts, model.ID.ValueString()
}

func flattenTeamMembership(userID, teamID, role string) resourceTeamMembershipModel {
	model := resourceTeamMembershipModel{
		ID:     flattenTeamMembershipID(userID, teamID),
		UserID: types.StringValue(userID),
		TeamID: types.StringValue(teamID),
		Role:   types.StringValue(role),
	}
	return model
}

func flattenTeamMembershipID(userID, teamID string) types.String {
	return types.StringValue(fmt.Sprintf("%s:%s", userID, teamID))
}
