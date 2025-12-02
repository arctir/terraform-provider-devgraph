package provider

import (
	"context"
	"fmt"
	"net/url"

	v1 "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &OAuthServiceResource{}
	_ resource.ResourceWithConfigure   = &OAuthServiceResource{}
	_ resource.ResourceWithImportState = &OAuthServiceResource{}
)

func NewOAuthServiceResource() resource.Resource {
	return &OAuthServiceResource{}
}

type OAuthServiceResource struct {
	client *v1.Client
}

type OAuthServiceResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	DisplayName         types.String `tfsdk:"display_name"`
	Description         types.String `tfsdk:"description"`
	ClientID            types.String `tfsdk:"client_id"`
	ClientSecret        types.String `tfsdk:"client_secret"`
	AuthorizationURL    types.String `tfsdk:"authorization_url"`
	TokenURL            types.String `tfsdk:"token_url"`
	UserinfoURL         types.String `tfsdk:"userinfo_url"`
	DefaultScopes       types.List   `tfsdk:"default_scopes"`
	SupportedGrantTypes types.List   `tfsdk:"supported_grant_types"`
	IsActive            types.Bool   `tfsdk:"is_active"`
	IconURL             types.String `tfsdk:"icon_url"`
	HomepageURL         types.String `tfsdk:"homepage_url"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
}

func (r *OAuthServiceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oauth_service"
}

func (r *OAuthServiceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an OAuth Service configuration in Devgraph.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the OAuth service.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the OAuth service (used as identifier).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"display_name": schema.StringAttribute{
				Description: "The display name of the OAuth service.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the OAuth service.",
				Optional:    true,
			},
			"client_id": schema.StringAttribute{
				Description: "The OAuth client ID.",
				Required:    true,
			},
			"client_secret": schema.StringAttribute{
				Description: "The OAuth client secret.",
				Required:    true,
				Sensitive:   true,
			},
			"authorization_url": schema.StringAttribute{
				Description: "The OAuth authorization endpoint URL.",
				Required:    true,
			},
			"token_url": schema.StringAttribute{
				Description: "The OAuth token endpoint URL.",
				Required:    true,
			},
			"userinfo_url": schema.StringAttribute{
				Description: "The OAuth userinfo endpoint URL.",
				Optional:    true,
			},
			"default_scopes": schema.ListAttribute{
				Description: "Default OAuth scopes to request.",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"supported_grant_types": schema.ListAttribute{
				Description: "Supported OAuth grant types.",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether the OAuth service is active.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"icon_url": schema.StringAttribute{
				Description: "URL to the service icon.",
				Optional:    true,
			},
			"homepage_url": schema.StringAttribute{
				Description: "URL to the service homepage.",
				Optional:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the OAuth service was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the OAuth service was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *OAuthServiceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*v1.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *v1.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *OAuthServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OAuthServiceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse required URLs
	authURL, err := url.Parse(plan.AuthorizationURL.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid authorization URL", err.Error())
		return
	}

	tokenURL, err := url.Parse(plan.TokenURL.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid token URL", err.Error())
		return
	}

	// Convert supported_grant_types to []string
	var supportedGrantTypes []string
	if !plan.SupportedGrantTypes.IsNull() && !plan.SupportedGrantTypes.IsUnknown() {
		diags = plan.SupportedGrantTypes.ElementsAs(ctx, &supportedGrantTypes, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Build the create request
	createReq := v1.OAuthServiceCreate{
		Name:                plan.Name.ValueString(),
		DisplayName:         plan.DisplayName.ValueString(),
		ClientID:            plan.ClientID.ValueString(),
		ClientSecret:        plan.ClientSecret.ValueString(),
		AuthorizationURL:    *authURL,
		TokenURL:            *tokenURL,
		SupportedGrantTypes: supportedGrantTypes,
	}

	// Add optional fields
	if !plan.Description.IsNull() {
		desc := plan.Description.ValueString()
		createReq.Description = v1.NewOptNilString(desc)
	}

	if !plan.DefaultScopes.IsNull() && !plan.DefaultScopes.IsUnknown() {
		var defaultScopes []string
		diags = plan.DefaultScopes.ElementsAs(ctx, &defaultScopes, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.DefaultScopes = v1.NewOptNilStringArray(defaultScopes)
	}

	if !plan.UserinfoURL.IsNull() {
		userInfoURL, err := url.Parse(plan.UserinfoURL.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Invalid userinfo URL", err.Error())
			return
		}
		createReq.UserinfoURL = v1.NewOptNilURI(*userInfoURL)
	}

	if !plan.IsActive.IsNull() {
		createReq.IsActive = v1.NewOptBool(plan.IsActive.ValueBool())
	}

	if !plan.IconURL.IsNull() {
		iconURL, err := url.Parse(plan.IconURL.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Invalid icon URL", err.Error())
			return
		}
		createReq.IconURL = v1.NewOptNilURI(*iconURL)
	}

	if !plan.HomepageURL.IsNull() {
		homepageURL, err := url.Parse(plan.HomepageURL.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Invalid homepage URL", err.Error())
			return
		}
		createReq.HomepageURL = v1.NewOptNilURI(*homepageURL)
	}

	resultInterface, err := r.client.CreateOAuthService(ctx, &createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating OAuth service",
			"Could not create OAuth service: "+err.Error(),
		)
		return
	}

	// Type assert the response
	result, ok := resultInterface.(*v1.OAuthServiceResponse)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected response type",
			fmt.Sprintf("Expected *v1.OAuthServiceResponse, got: %T", resultInterface),
		)
		return
	}

	// Update state with created resource
	plan.ID = types.StringValue(result.ID.String())
	// Keep the plan.Name as requested - don't overwrite with API response
	plan.DisplayName = types.StringValue(result.DisplayName)
	plan.AuthorizationURL = types.StringValue(result.AuthorizationURL)
	plan.TokenURL = types.StringValue(result.TokenURL)
	plan.IsActive = types.BoolValue(result.IsActive)
	plan.CreatedAt = types.StringValue(result.CreatedAt.String())
	plan.UpdatedAt = types.StringValue(result.UpdatedAt.String())

	if !result.Description.Null {
		plan.Description = types.StringValue(result.Description.Value)
	} else {
		plan.Description = types.StringNull()
	}

	if !result.UserinfoURL.Null {
		plan.UserinfoURL = types.StringValue(result.UserinfoURL.Value)
	} else {
		plan.UserinfoURL = types.StringNull()
	}

	if !result.IconURL.Null {
		plan.IconURL = types.StringValue(result.IconURL.Value)
	} else {
		plan.IconURL = types.StringNull()
	}

	if !result.HomepageURL.Null {
		plan.HomepageURL = types.StringValue(result.HomepageURL.Value)
	} else {
		plan.HomepageURL = types.StringNull()
	}

	// Convert scopes back to list
	scopeValues := make([]attr.Value, len(result.DefaultScopes))
	for i, scope := range result.DefaultScopes {
		scopeValues[i] = types.StringValue(scope)
	}
	plan.DefaultScopes = types.ListValueMust(types.StringType, scopeValues)

	// Convert grant types back to list
	grantTypeValues := make([]attr.Value, len(result.SupportedGrantTypes))
	for i, grantType := range result.SupportedGrantTypes {
		grantTypeValues[i] = types.StringValue(grantType)
	}
	plan.SupportedGrantTypes = types.ListValueMust(types.StringType, grantTypeValues)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *OAuthServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OAuthServiceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid OAuth Service ID", err.Error())
		return
	}

	resultInterface, err := r.client.GetOAuthService(ctx, v1.GetOAuthServiceParams{
		ServiceID: serviceID,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading OAuth service",
			"Could not read OAuth service ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Check if resource was not found (deleted outside Terraform)
	if _, ok := resultInterface.(*v1.GetOAuthServiceNotFound); ok {
		// Resource was deleted outside Terraform, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Type assert the response
	result, ok := resultInterface.(*v1.OAuthServiceResponse)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected response type",
			fmt.Sprintf("Expected *v1.OAuthServiceResponse, got: %T", resultInterface),
		)
		return
	}

	// Update state
	// Keep the state.Name - don't overwrite with API response
	state.DisplayName = types.StringValue(result.DisplayName)
	state.AuthorizationURL = types.StringValue(result.AuthorizationURL)
	state.TokenURL = types.StringValue(result.TokenURL)
	state.IsActive = types.BoolValue(result.IsActive)
	state.CreatedAt = types.StringValue(result.CreatedAt.String())
	state.UpdatedAt = types.StringValue(result.UpdatedAt.String())

	if !result.Description.Null {
		state.Description = types.StringValue(result.Description.Value)
	} else {
		state.Description = types.StringNull()
	}

	if !result.UserinfoURL.Null {
		state.UserinfoURL = types.StringValue(result.UserinfoURL.Value)
	} else {
		state.UserinfoURL = types.StringNull()
	}

	if !result.IconURL.Null {
		state.IconURL = types.StringValue(result.IconURL.Value)
	} else {
		state.IconURL = types.StringNull()
	}

	if !result.HomepageURL.Null {
		state.HomepageURL = types.StringValue(result.HomepageURL.Value)
	} else {
		state.HomepageURL = types.StringNull()
	}

	// Convert scopes back to list
	scopeValues := make([]attr.Value, len(result.DefaultScopes))
	for i, scope := range result.DefaultScopes {
		scopeValues[i] = types.StringValue(scope)
	}
	state.DefaultScopes = types.ListValueMust(types.StringType, scopeValues)

	// Convert grant types back to list
	grantTypeValues := make([]attr.Value, len(result.SupportedGrantTypes))
	for i, grantType := range result.SupportedGrantTypes {
		grantTypeValues[i] = types.StringValue(grantType)
	}
	state.SupportedGrantTypes = types.ListValueMust(types.StringType, grantTypeValues)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *OAuthServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OAuthServiceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID, err := uuid.Parse(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid OAuth Service ID", err.Error())
		return
	}

	// Build the update request
	updateReq := v1.OAuthServiceUpdate{}

	if !plan.DisplayName.IsNull() {
		updateReq.DisplayName = v1.NewOptNilString(plan.DisplayName.ValueString())
	}

	if !plan.ClientID.IsNull() {
		updateReq.ClientID = v1.NewOptNilString(plan.ClientID.ValueString())
	}

	if !plan.ClientSecret.IsNull() {
		updateReq.ClientSecret = v1.NewOptNilString(plan.ClientSecret.ValueString())
	}

	if !plan.AuthorizationURL.IsNull() {
		authURL, err := url.Parse(plan.AuthorizationURL.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Invalid authorization URL", err.Error())
			return
		}
		updateReq.AuthorizationURL = v1.NewOptNilURI(*authURL)
	}

	if !plan.TokenURL.IsNull() {
		tokenURL, err := url.Parse(plan.TokenURL.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Invalid token URL", err.Error())
			return
		}
		updateReq.TokenURL = v1.NewOptNilURI(*tokenURL)
	}

	if !plan.Description.IsNull() {
		updateReq.Description = v1.NewOptNilString(plan.Description.ValueString())
	}

	if !plan.DefaultScopes.IsNull() && !plan.DefaultScopes.IsUnknown() {
		var defaultScopes []string
		diags = plan.DefaultScopes.ElementsAs(ctx, &defaultScopes, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		updateReq.DefaultScopes = v1.NewOptNilStringArray(defaultScopes)
	}

	if !plan.SupportedGrantTypes.IsNull() && !plan.SupportedGrantTypes.IsUnknown() {
		var supportedGrantTypes []string
		diags = plan.SupportedGrantTypes.ElementsAs(ctx, &supportedGrantTypes, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		updateReq.SupportedGrantTypes = v1.NewOptNilStringArray(supportedGrantTypes)
	}

	if !plan.UserinfoURL.IsNull() {
		userInfoURL, err := url.Parse(plan.UserinfoURL.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Invalid userinfo URL", err.Error())
			return
		}
		updateReq.UserinfoURL = v1.NewOptNilURI(*userInfoURL)
	}

	if !plan.IsActive.IsNull() {
		updateReq.IsActive = v1.NewOptNilBool(plan.IsActive.ValueBool())
	}

	if !plan.IconURL.IsNull() {
		iconURL, err := url.Parse(plan.IconURL.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Invalid icon URL", err.Error())
			return
		}
		updateReq.IconURL = v1.NewOptNilURI(*iconURL)
	}

	if !plan.HomepageURL.IsNull() {
		homepageURL, err := url.Parse(plan.HomepageURL.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Invalid homepage URL", err.Error())
			return
		}
		updateReq.HomepageURL = v1.NewOptNilURI(*homepageURL)
	}

	resultInterface, err := r.client.UpdateOAuthService(ctx, &updateReq, v1.UpdateOAuthServiceParams{
		ServiceID: serviceID,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating OAuth service",
			"Could not update OAuth service: "+err.Error(),
		)
		return
	}

	// Type assert the response
	result, ok := resultInterface.(*v1.OAuthServiceResponse)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected response type",
			fmt.Sprintf("Expected *v1.OAuthServiceResponse, got: %T", resultInterface),
		)
		return
	}

	// Update state with updated resource
	// Keep the plan.Name - don't overwrite with API response
	plan.DisplayName = types.StringValue(result.DisplayName)
	plan.AuthorizationURL = types.StringValue(result.AuthorizationURL)
	plan.TokenURL = types.StringValue(result.TokenURL)
	plan.IsActive = types.BoolValue(result.IsActive)
	plan.UpdatedAt = types.StringValue(result.UpdatedAt.String())

	if !result.Description.Null {
		plan.Description = types.StringValue(result.Description.Value)
	} else {
		plan.Description = types.StringNull()
	}

	if !result.UserinfoURL.Null {
		plan.UserinfoURL = types.StringValue(result.UserinfoURL.Value)
	} else {
		plan.UserinfoURL = types.StringNull()
	}

	if !result.IconURL.Null {
		plan.IconURL = types.StringValue(result.IconURL.Value)
	} else {
		plan.IconURL = types.StringNull()
	}

	if !result.HomepageURL.Null {
		plan.HomepageURL = types.StringValue(result.HomepageURL.Value)
	} else {
		plan.HomepageURL = types.StringNull()
	}

	// Convert scopes back to list
	scopeValues := make([]attr.Value, len(result.DefaultScopes))
	for i, scope := range result.DefaultScopes {
		scopeValues[i] = types.StringValue(scope)
	}
	plan.DefaultScopes = types.ListValueMust(types.StringType, scopeValues)

	// Convert grant types back to list
	grantTypeValues := make([]attr.Value, len(result.SupportedGrantTypes))
	for i, grantType := range result.SupportedGrantTypes {
		grantTypeValues[i] = types.StringValue(grantType)
	}
	plan.SupportedGrantTypes = types.ListValueMust(types.StringType, grantTypeValues)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *OAuthServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OAuthServiceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid OAuth Service ID", err.Error())
		return
	}

	_, err = r.client.DeleteOAuthService(ctx, v1.DeleteOAuthServiceParams{
		ServiceID: serviceID,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting OAuth service",
			"Could not delete OAuth service: "+err.Error(),
		)
		return
	}
}

func (r *OAuthServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
