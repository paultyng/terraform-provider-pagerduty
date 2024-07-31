package pagerduty

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/PagerDuty/terraform-provider-pagerduty/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

type resourceUserNotificationRule struct{ client *pagerduty.Client }

var (
	_ resource.ResourceWithConfigure   = (*resourceUserNotificationRule)(nil)
	_ resource.ResourceWithImportState = (*resourceUserNotificationRule)(nil)
)

func (r *resourceUserNotificationRule) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "pagerduty_user_notification_rule"
}

func (r *resourceUserNotificationRule) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"user_id":                schema.StringAttribute{Required: true},
			"start_delay_in_minutes": schema.Int64Attribute{Required: true},
			"urgency": schema.StringAttribute{
				Required:   true,
				Validators: []validator.String{stringvalidator.OneOf("high", "low")},
			},
		},
		Blocks: map[string]schema.Block{
			"contact_method": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Required: true,
						Validators: []validator.String{
							stringvalidator.OneOf(
								"email_contact_method",
								"phone_contact_method",
								"push_notification_contact_method",
								"sms_contact_method",
							),
						},
					},
					"id": schema.StringAttribute{Required: true},
				},
			},
		},
	}
}

func (r *resourceUserNotificationRule) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model resourceUserNotificationRuleModel
	var userID types.String

	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("user_id"), &userID)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan := buildPagerdutyUserNotificationRule(ctx, &model, &resp.Diagnostics)
	log.Printf("[INFO] Creating PagerDuty user notification rule for %s", userID)

	response, err := r.client.CreateUserNotificationRuleWithContext(ctx, userID.ValueString(), plan)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error creating PagerDuty user notification rule for %s", userID),
			err.Error(),
		)
		return
	}

	model, err = requestGetUserNotificationRule(ctx, r.client, userID.ValueString(), response.ID, true, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading PagerDuty user notification rule for %s", userID),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *resourceUserNotificationRule) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var id types.String
	var userID types.String

	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("user_id"), &userID)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[INFO] Reading PagerDuty user notification rule %s", id)

	state, err := requestGetUserNotificationRule(ctx, r.client, userID.ValueString(), id.ValueString(), false, &resp.Diagnostics)
	if err != nil {
		if util.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading PagerDuty user notification rule %s", id),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceUserNotificationRule) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model resourceUserNotificationRuleModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan := buildPagerdutyUserNotificationRule(ctx, &model, &resp.Diagnostics)
	if plan.ID == "" {
		var id string
		req.State.GetAttribute(ctx, path.Root("id"), &id)
		plan.ID = id
	}

	var userIDState types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("user_id"), &userIDState)...)
	if resp.Diagnostics.HasError() {
		return
	}
	userID := userIDState.ValueString()

	if userID == "" {
		var userIDPlan types.String
		resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("user_id"), &userIDPlan)...)
		if resp.Diagnostics.HasError() {
			return
		}
		userID = userIDPlan.ValueString()
	}
	log.Printf("[INFO] Updating PagerDuty user notification rule %s", plan.ID)

	_, err := r.client.UpdateUserNotificationRuleWithContext(ctx, userID, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error updating PagerDuty user notification rule %s", plan.ID),
			err.Error(),
		)
		return
	}

	model, err = requestGetUserNotificationRule(ctx, r.client, userID, plan.ID, false, &resp.Diagnostics)
	if err != nil {
		if util.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading PagerDuty user notification rule %s", plan.ID),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *resourceUserNotificationRule) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var id types.String
	var userID types.String

	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("user_id"), &userID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[INFO] Deleting PagerDuty user notification rule %s", id)

	err := r.client.DeleteUserNotificationRuleWithContext(ctx, userID.ValueString(), id.ValueString())
	if err != nil && !util.IsNotFoundError(err) {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error deleting PagerDuty user notification rule %s", id),
			err.Error(),
		)
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *resourceUserNotificationRule) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(ConfigurePagerdutyClient(&r.client, req.ProviderData)...)
}

func (r *resourceUserNotificationRule) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ids := strings.Split(req.ID, ":")
	if len(ids) != 2 {
		resp.Diagnostics.AddError(
			"Error importing pagerduty_user_notification_rule",
			"Expecting an ID formed as '<user_id>.<notification_rule_id>'",
		)
		return
	}
	uid, id := ids[0], ids[1]

	model, err := requestGetUserNotificationRule(ctx, r.client, uid, id, true, &resp.Diagnostics)
	if err != nil {
		if util.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error importing PagerDuty user notification rule %s", id),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

type resourceUserNotificationRuleModel struct {
	ID                  types.String `tfsdk:"id"`
	UserID              types.String `tfsdk:"user_id"`
	StartDelayInMinutes types.Int64  `tfsdk:"start_delay_in_minutes"`
	Urgency             types.String `tfsdk:"urgency"`
	ContactMethod       types.Object `tfsdk:"contact_method"`
}

func requestGetUserNotificationRule(ctx context.Context, client *pagerduty.Client, userID string, id string, retryNotFound bool, diags *diag.Diagnostics) (resourceUserNotificationRuleModel, error) {
	var model resourceUserNotificationRuleModel

	err := retry.RetryContext(ctx, 2*time.Minute, func() *retry.RetryError {
		notificationRule, err := client.GetUserNotificationRuleWithContext(ctx, userID, id)
		if err != nil {
			if util.IsBadRequestError(err) {
				return retry.NonRetryableError(err)
			}
			if !retryNotFound && util.IsNotFoundError(err) {
				return retry.NonRetryableError(err)
			}
			return retry.RetryableError(err)
		}
		model = flattenUserNotificationRule(notificationRule, userID)
		return nil
	})

	return model, err
}

func buildPagerdutyUserNotificationRule(ctx context.Context, model *resourceUserNotificationRuleModel, diags *diag.Diagnostics) pagerduty.NotificationRule {
	return pagerduty.NotificationRule{
		ID:                  model.ID.ValueString(),
		ContactMethod:       buildPagerDutyContactMethodReference(ctx, model.ContactMethod, diags),
		StartDelayInMinutes: uint(model.StartDelayInMinutes.ValueInt64()),
		Type:                "assignment_notification_rule",
		Urgency:             model.Urgency.ValueString(),
	}
}

func buildPagerDutyContactMethodReference(ctx context.Context, contactMethod types.Object, diags *diag.Diagnostics) pagerduty.ContactMethod {
	var target struct {
		ID   types.String `tfsdk:"id"`
		Type types.String `tfsdk:"type"`
	}

	d := contactMethod.As(ctx, &target, basetypes.ObjectAsOptions{})
	diags.Append(d...)

	return pagerduty.ContactMethod{
		ID:   target.ID.ValueString(),
		Type: target.Type.ValueString(),
	}
}

func flattenUserNotificationRule(response *pagerduty.NotificationRule, userID string) resourceUserNotificationRuleModel {
	contactMethodObjectType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"id":   types.StringType,
		"type": types.StringType,
	}}
	model := resourceUserNotificationRuleModel{
		ID:                  types.StringValue(response.ID),
		StartDelayInMinutes: types.Int64Value(int64(response.StartDelayInMinutes)),
		Urgency:             types.StringValue(response.Urgency),
		UserID:              types.StringValue(userID),
		ContactMethod: types.ObjectValueMust(contactMethodObjectType.AttrTypes, map[string]attr.Value{
			"id":   types.StringValue(response.ContactMethod.ID),
			"type": types.StringValue(response.ContactMethod.Type),
		}),
	}
	return model
}

/*

func resourcePagerDutyUserNotificationRuleImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client, err := meta.(*Config).Client()
	if err != nil {
		return []*schema.ResourceData{}, err
	}

	ids := strings.Split(d.Id(), ":")

	if len(ids) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Error importing pagerduty_user_notification_rule. Expecting an ID formed as '<user_id>.<notification_rule_id>'")
	}
	uid, id := ids[0], ids[1]

	_, _, err = client.Users.GetNotificationRule(uid, id)
	if err != nil {
		return []*schema.ResourceData{}, err
	}

	d.SetId(id)
	d.Set("user_id", uid)

	return []*schema.ResourceData{d}, nil
}

func expandContactMethod(v interface{}) (*pagerduty.ContactMethodReference, error) {
        cm := v.(map[string]interface{})
        if _, ok := cm["id"]; !ok {
                return nil, fmt.Errorf("the `id` attribute of `contact_method` is required")
        }
        if t, ok := cm["type"]; !ok {
                return nil, fmt.Errorf("the `type` attribute of `contact_method` is required")
        } else {
                switch t {
                case "email_contact_method":
                case "phone_contact_method":
                case "push_notification_contact_method":
                case "sms_contact_method":
                        // Valid
                default:
                        return nil, fmt.Errorf("the `type` attribute of `contact_method` must be one of `email_contact_method`, `phone_contact_method`, `push_notification_contact_method` or `sms_co>
                }
        }
        contactMethod := &pagerduty.ContactMethodReference{
                ID:   cm["id"].(string),
                Type: cm["type"].(string),
        }
        return contactMethod, nil
}

*/
