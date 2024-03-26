package pagerduty

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Provider struct {
	client *pagerduty.Client
}

func (p *Provider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "pagerduty"
}

func (p *Provider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	useAppOauthScopedTokenBlock := schema.ListNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"pd_client_id":     schema.StringAttribute{Optional: true},
				"pd_client_secret": schema.StringAttribute{Optional: true},
				"pd_subdomain":     schema.StringAttribute{Optional: true},
			},
		},
	}
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_url_override":            schema.StringAttribute{Optional: true},
			"service_region":              schema.StringAttribute{Optional: true},
			"skip_credentials_validation": schema.BoolAttribute{Optional: true},
			"token":                       schema.StringAttribute{Optional: true},
			"user_token":                  schema.StringAttribute{Optional: true},
			"insecure_tls":                schema.BoolAttribute{Optional: true},
		},
		Blocks: map[string]schema.Block{
			"use_app_oauth_scoped_token": useAppOauthScopedTokenBlock,
		},
	}
}

func (p *Provider) DataSources(_ context.Context) [](func() datasource.DataSource) {
	return [](func() datasource.DataSource){
		func() datasource.DataSource { return &dataSourceBusinessService{} },
		func() datasource.DataSource { return &dataSourceEscalationPolicy{} },
		func() datasource.DataSource { return &dataSourceExtensionSchema{} },
		func() datasource.DataSource { return &dataSourceIntegration{} },
		func() datasource.DataSource { return &dataSourceLicenses{} },
		func() datasource.DataSource { return &dataSourceLicense{} },
		func() datasource.DataSource { return &dataSourcePriority{} },
		func() datasource.DataSource { return &dataSourceService{} },
		func() datasource.DataSource { return &dataSourceStandardsResourceScores{} },
		func() datasource.DataSource { return &dataSourceStandardsResourcesScores{} },
		func() datasource.DataSource { return &dataSourceStandards{} },
		func() datasource.DataSource { return &dataSourceTag{} },
		func() datasource.DataSource { return &dataSourceUsers{} },
		func() datasource.DataSource { return &dataSourceUser{} },
		func() datasource.DataSource { return &dataSourceVendor{} },
	}
}

func (p *Provider) Resources(_ context.Context) [](func() resource.Resource) {
	return [](func() resource.Resource){
		func() resource.Resource { return &resourceAddon{} },
		func() resource.Resource { return &resourceBusinessService{} },
		func() resource.Resource { return &resourceExtensionServiceNow{} },
		func() resource.Resource { return &resourceExtension{} },
		func() resource.Resource { return &resourceServiceDependency{} },
		func() resource.Resource { return &resourceTagAssignment{} },
		func() resource.Resource { return &resourceTag{} },
		func() resource.Resource { return &resourceTeamMembership{} },
		func() resource.Resource { return &resourceTeam{} },
		func() resource.Resource { return &resourceUserHandoffNotificationRule{} },
		func() resource.Resource { return &resourceUserNotificationRule{} },
	}
}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var args providerArguments
	resp.Diagnostics.Append(req.Config.Get(ctx, &args)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceRegion := args.ServiceRegion.ValueString()
	if serviceRegion == "" {
		if v, ok := os.LookupEnv("PAGERDUTY_SERVICE_REGION"); ok && v != "" {
			serviceRegion = v
		} else {
			serviceRegion = "us"
		}
	}

	regionAPIURL := ""
	if serviceRegion != "us" {
		regionAPIURL = serviceRegion + "."
	}

	skipCredentialsValidation := args.SkipCredentialsValidation.Equal(types.BoolValue(true))
	insecureTls := args.InsecureTls.Equal(types.BoolValue(true))

	config := Config{
		APIURL:              "https://api." + regionAPIURL + "pagerduty.com",
		AppURL:              "https://app." + regionAPIURL + "pagerduty.com",
		SkipCredsValidation: skipCredentialsValidation,
		Token:               args.Token.ValueString(),
		UserToken:           args.UserToken.ValueString(),
		TerraformVersion:    req.TerraformVersion,
		APIURLOverride:      args.APIURLOverride.ValueString(),
		ServiceRegion:       serviceRegion,
		InsecureTls:         insecureTls,
	}

	if !args.UseAppOauthScopedToken.IsNull() {
		blockList := []UseAppOauthScopedToken{}
		resp.Diagnostics.Append(args.UseAppOauthScopedToken.ElementsAs(ctx, &blockList, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		config.AppOauthScopedToken = &AppOauthScopedToken{
			ClientID:     blockList[0].PdClientID.ValueString(),
			ClientSecret: blockList[0].PdClientSecret.ValueString(),
			Subdomain:    blockList[0].PdSubdomain.ValueString(),
		}
	}

	if args.UseAppOauthScopedToken.IsNull() {
		if config.Token == "" {
			config.Token = os.Getenv("PAGERDUTY_TOKEN")
		}
		if config.UserToken == "" {
			config.UserToken = os.Getenv("PAGERDUTY_USER_TOKEN")
		}
	} else {
		if config.AppOauthScopedToken.ClientID == "" {
			config.AppOauthScopedToken.ClientID = os.Getenv("PAGERDUTY_CLIENT_ID")
		}
		if config.AppOauthScopedToken.ClientSecret == "" {
			config.AppOauthScopedToken.ClientSecret = os.Getenv("PAGERDUTY_CLIENT_SECRET")
		}
		if config.AppOauthScopedToken.Subdomain == "" {
			config.AppOauthScopedToken.Subdomain = os.Getenv("PAGERDUTY_SUBDOMAIN")
		}
	}

	if config.AppOauthScopedToken != nil {
		// While doing migration to terraform plugin framework, because
		// of a limitation of the provider mux
		// https://github.com/hashicorp/terraform-plugin-framework/issues/539
		// We had to define pd_client_id, pd_client_secret, and pd_subdomain
		// as Optional and manually check its presence here.
		li := []string{}
		if config.AppOauthScopedToken.ClientID == "" {
			li = append(li, "pd_client_id")
		}
		if config.AppOauthScopedToken.ClientSecret == "" {
			li = append(li, "pd_client_secret")
		}
		if config.AppOauthScopedToken.Subdomain == "" {
			li = append(li, "pd_subdomain")
		}
		if len(li) > 0 {
			resp.Diagnostics.AddError(
				fmt.Sprintf(`Missing required arguments: "%v"`, strings.Join(li, `", "`)),
				"Despite being defined as Optional, its value is required and no definition was found.",
			)
			return
		}
	}

	log.Println("[INFO] Initializing PagerDuty plugin client")

	client, err := config.Client(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Cannot obtain plugin client", err.Error())
	}
	p.client = client
	resp.DataSourceData = client
	resp.ResourceData = client
}

type UseAppOauthScopedToken struct {
	PdClientID     types.String `tfsdk:"pd_client_id"`
	PdClientSecret types.String `tfsdk:"pd_client_secret"`
	PdSubdomain    types.String `tfsdk:"pd_subdomain"`
}

type providerArguments struct {
	Token                     types.String `tfsdk:"token"`
	UserToken                 types.String `tfsdk:"user_token"`
	SkipCredentialsValidation types.Bool   `tfsdk:"skip_credentials_validation"`
	ServiceRegion             types.String `tfsdk:"service_region"`
	APIURLOverride            types.String `tfsdk:"api_url_override"`
	UseAppOauthScopedToken    types.List   `tfsdk:"use_app_oauth_scoped_token"`
	InsecureTls               types.Bool   `tfsdk:"insecure_tls"`
}

type SchemaGetter interface {
	Get(context.Context, interface{}) diag.Diagnostics
	GetAttribute(context.Context, path.Path, interface{}) diag.Diagnostics
}

func extractString(ctx context.Context, schema SchemaGetter, name string, diags *diag.Diagnostics) *string {
	var s types.String
	d := schema.GetAttribute(ctx, path.Root(name), &s)
	diags.Append(d...)
	return s.ValueStringPointer()
}

// Helper constant used to have a semantically meaningful value in function calls
const RetryNotFound = true
