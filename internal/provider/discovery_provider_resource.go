package provider

import (
	"context"
	"encoding/json"
	"fmt"

	v1 "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
	"github.com/go-faster/jx"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &DiscoveryProviderResource{}
	_ resource.ResourceWithConfigure   = &DiscoveryProviderResource{}
	_ resource.ResourceWithImportState = &DiscoveryProviderResource{}
)

func NewDiscoveryProviderResource() resource.Resource {
	return &DiscoveryProviderResource{}
}

type DiscoveryProviderResource struct {
	client *v1.Client
}

type DiscoveryProviderResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	ProviderType types.String `tfsdk:"provider_type"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	Interval     types.Int64  `tfsdk:"interval"`
	Config       types.String `tfsdk:"config"`
}

func (r *DiscoveryProviderResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_discovery_provider"
}

func (r *DiscoveryProviderResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a discovery provider in Devgraph. Discovery providers continuously discover entities and relations from external systems (GitHub, GitLab, Argo, etc.).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the discovery provider.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Human-readable name for this provider instance (e.g., 'GitHub Production').",
				Required:    true,
			},
			"provider_type": schema.StringAttribute{
				Description: "Type of provider (github, gitlab, argo, vercel, docker, file, fossa, meta).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether this provider is active and should run discovery.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"interval": schema.Int64Attribute{
				Description: "How often to run discovery, in seconds (minimum 60).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(300),
			},
			"config": schema.StringAttribute{
				Description: "Provider configuration as JSON string. The configuration schema depends on the provider_type. Sensitive values (tokens, API keys) will be encrypted.",
				Required:    true,
				Sensitive:   true,
			},
		},
	}
}

func (r *DiscoveryProviderResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DiscoveryProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DiscoveryProviderResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse config JSON into map[string]jx.Raw
	configJSON := []byte(plan.Config.ValueString())

	// Validate it's valid JSON first
	var validateMap map[string]interface{}
	if err := json.Unmarshal(configJSON, &validateMap); err != nil {
		resp.Diagnostics.AddError(
			"Invalid Config JSON",
			"Could not parse config as JSON: "+err.Error(),
		)
		return
	}

	// Convert to map[string]jx.Raw
	configMap := make(v1.ConfiguredProviderCreateConfig)
	for key, value := range validateMap {
		valueJSON, err := json.Marshal(value)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error encoding config value",
				fmt.Sprintf("Could not encode value for key %s: %v", key, err),
			)
			return
		}
		configMap[key] = jx.Raw(valueJSON)
	}

	// Build create request
	createReq := v1.ConfiguredProviderCreate{
		Name:         plan.Name.ValueString(),
		ProviderType: plan.ProviderType.ValueString(),
		Config:       configMap,
	}

	if !plan.Enabled.IsNull() {
		enabled := plan.Enabled.ValueBool()
		createReq.SetEnabled(v1.NewOptBool(enabled))
	}

	if !plan.Interval.IsNull() {
		interval := int(plan.Interval.ValueInt64())
		createReq.SetInterval(v1.NewOptInt(interval))
	}

	// Create provider
	res, err := r.client.CreateConfiguredProvider(ctx, &createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating discovery provider",
			"Could not create discovery provider: "+err.Error(),
		)
		return
	}

	// Type assert the response - handle different response types
	switch result := res.(type) {
	case *v1.ConfiguredProviderResponse:
		// Success case
		plan.ID = types.StringValue(result.ID.String())
		plan.Name = types.StringValue(result.Name)
		plan.ProviderType = types.StringValue(result.ProviderType)
		plan.Enabled = types.BoolValue(result.Enabled)
		plan.Interval = types.Int64Value(int64(result.Interval))
	case *v1.CreateConfiguredProviderNotFound:
		resp.Diagnostics.AddError(
			"Provider type not found",
			fmt.Sprintf("The provider type '%s' was not found. Check that it's a valid provider type (github, gitlab, argo, vercel, docker, file, fossa).", plan.ProviderType.ValueString()),
		)
		return
	default:
		resp.Diagnostics.AddError(
			"Unexpected response type",
			fmt.Sprintf("Expected *v1.ConfiguredProviderResponse, got: %T", res),
		)
		return
	}

	// Keep the original config in state (not the masked one from response)
	// This allows Terraform to detect config changes

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *DiscoveryProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DiscoveryProviderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse UUID
	providerID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid provider ID",
			"Could not parse provider ID as UUID: "+err.Error(),
		)
		return
	}

	// Get provider
	res, err := r.client.GetConfiguredProvider(ctx, v1.GetConfiguredProviderParams{
		ProviderID: providerID,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading discovery provider",
			"Could not read discovery provider: "+err.Error(),
		)
		return
	}

	// Type assert the response - handle different response types
	switch result := res.(type) {
	case *v1.ConfiguredProviderResponse:
		// Success case - update state
		state.Name = types.StringValue(result.Name)
		state.ProviderType = types.StringValue(result.ProviderType)
		state.Enabled = types.BoolValue(result.Enabled)
		state.Interval = types.Int64Value(int64(result.Interval))
	case *v1.GetConfiguredProviderNotFound:
		// Resource doesn't exist - remove from state
		resp.State.RemoveResource(ctx)
		return
	default:
		resp.Diagnostics.AddError(
			"Unexpected response type",
			fmt.Sprintf("Expected *v1.ConfiguredProviderResponse, got: %T", res),
		)
		return
	}

	// Keep the config from state since the API returns masked secrets
	// This prevents Terraform from thinking the config has changed

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *DiscoveryProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DiscoveryProviderResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state DiscoveryProviderResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse UUID
	providerID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid provider ID",
			"Could not parse provider ID as UUID: "+err.Error(),
		)
		return
	}

	// Build update request
	updateReq := v1.ConfiguredProviderUpdate{}

	if !plan.Name.Equal(state.Name) {
		name := plan.Name.ValueString()
		updateReq.SetName(v1.NewOptNilString(name))
	}

	if !plan.Enabled.Equal(state.Enabled) {
		enabled := plan.Enabled.ValueBool()
		updateReq.SetEnabled(v1.NewOptNilBool(enabled))
	}

	if !plan.Interval.Equal(state.Interval) {
		interval := int(plan.Interval.ValueInt64())
		updateReq.SetInterval(v1.NewOptNilInt(interval))
	}

	if !plan.Config.Equal(state.Config) {
		// Parse config JSON
		configJSON := []byte(plan.Config.ValueString())

		// Validate it's valid JSON first
		var validateMap map[string]interface{}
		if err := json.Unmarshal(configJSON, &validateMap); err != nil {
			resp.Diagnostics.AddError(
				"Invalid Config JSON",
				"Could not parse config as JSON: "+err.Error(),
			)
			return
		}

		// Convert to map[string]jx.Raw
		configMap := make(v1.ConfiguredProviderUpdateConfig)
		for key, value := range validateMap {
			valueJSON, err := json.Marshal(value)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error encoding config value",
					fmt.Sprintf("Could not encode value for key %s: %v", key, err),
				)
				return
			}
			configMap[key] = jx.Raw(valueJSON)
		}

		updateReq.SetConfig(v1.NewOptNilConfiguredProviderUpdateConfig(configMap))
	}

	// Update provider
	res, err := r.client.UpdateConfiguredProvider(ctx, &updateReq, v1.UpdateConfiguredProviderParams{
		ProviderID: providerID,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating discovery provider",
			"Could not update discovery provider: "+err.Error(),
		)
		return
	}

	// Type assert the response
	result, ok := res.(*v1.ConfiguredProviderResponse)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected response type",
			fmt.Sprintf("Expected *v1.ConfiguredProviderResponse, got: %T", res),
		)
		return
	}

	// Update state with updated resource
	plan.ID = types.StringValue(result.ID.String())
	plan.Name = types.StringValue(result.Name)
	plan.ProviderType = types.StringValue(result.ProviderType)
	plan.Enabled = types.BoolValue(result.Enabled)
	plan.Interval = types.Int64Value(int64(result.Interval))

	// Keep the config from plan since API returns masked secrets

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *DiscoveryProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DiscoveryProviderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse UUID
	providerID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid provider ID",
			"Could not parse provider ID as UUID: "+err.Error(),
		)
		return
	}

	// Delete provider
	_, err = r.client.DeleteConfiguredProvider(ctx, v1.DeleteConfiguredProviderParams{
		ProviderID: providerID,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting discovery provider",
			"Could not delete discovery provider: "+err.Error(),
		)
		return
	}
}

func (r *DiscoveryProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
