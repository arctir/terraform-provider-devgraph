package provider

import (
	"context"
	"fmt"

	v1 "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &MCPEndpointResource{}
	_ resource.ResourceWithConfigure   = &MCPEndpointResource{}
	_ resource.ResourceWithImportState = &MCPEndpointResource{}
)

func NewMCPEndpointResource() resource.Resource {
	return &MCPEndpointResource{}
}

type MCPEndpointResource struct {
	client *v1.Client
}

type MCPEndpointResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	URL               types.String `tfsdk:"url"`
	Description       types.String `tfsdk:"description"`
	Headers           types.Map    `tfsdk:"headers"`
	DevgraphAuth      types.Bool   `tfsdk:"devgraph_auth"`
	SupportsResources types.Bool   `tfsdk:"supports_resources"`
	OAuthServiceID    types.String `tfsdk:"oauth_service_id"`
	Immutable         types.Bool   `tfsdk:"immutable"`
	Active            types.Bool   `tfsdk:"active"`
	AllowedTools      types.List   `tfsdk:"allowed_tools"`
	DeniedTools       types.List   `tfsdk:"denied_tools"`
}

func (r *MCPEndpointResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mcp_endpoint"
}

func (r *MCPEndpointResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an MCP (Model Context Protocol) endpoint configuration in Devgraph.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the MCP endpoint.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the MCP endpoint.",
				Required:    true,
			},
			"url": schema.StringAttribute{
				Description: "The URL of the MCP endpoint.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the MCP endpoint.",
				Optional:    true,
			},
			"headers": schema.MapAttribute{
				Description: "Custom headers to send with requests to the MCP endpoint.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     mapdefault.StaticValue(types.MapValueMust(types.StringType, map[string]attr.Value{})),
			},
			"devgraph_auth": schema.BoolAttribute{
				Description: "Whether to use Devgraph authentication for this endpoint.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"supports_resources": schema.BoolAttribute{
				Description: "Whether this MCP endpoint supports resources.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"oauth_service_id": schema.StringAttribute{
				Description: "The OAuth service ID to use for authentication.",
				Optional:    true,
			},
			"immutable": schema.BoolAttribute{
				Description: "Whether this endpoint configuration is immutable.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"active": schema.BoolAttribute{
				Description: "Whether this MCP endpoint is active.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"allowed_tools": schema.ListAttribute{
				Description: "List of allowed tool names for this endpoint.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"denied_tools": schema.ListAttribute{
				Description: "List of denied tool names for this endpoint.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *MCPEndpointResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MCPEndpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MCPEndpointResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build headers map
	headers := make(map[string]string)
	if !plan.Headers.IsNull() {
		diags = plan.Headers.ElementsAs(ctx, &headers, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Build allowed tools list
	var allowedTools []string
	if !plan.AllowedTools.IsNull() {
		diags = plan.AllowedTools.ElementsAs(ctx, &allowedTools, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Build denied tools list
	var deniedTools []string
	if !plan.DeniedTools.IsNull() {
		diags = plan.DeniedTools.ElementsAs(ctx, &deniedTools, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Create MCP endpoint
	createReq := v1.MCPEndpointCreate{
		Name: plan.Name.ValueString(),
		URL:  plan.URL.ValueString(),
	}

	// Set optional description
	if !plan.Description.IsNull() {
		createReq.Description = v1.NewOptNilString(plan.Description.ValueString())
	}

	// Set headers if provided
	if len(headers) > 0 {
		createReq.Headers = v1.NewOptMCPEndpointCreateHeaders(v1.MCPEndpointCreateHeaders(headers))
	}

	// Set boolean fields
	createReq.DevgraphAuth = v1.NewOptBool(plan.DevgraphAuth.ValueBool())
	createReq.SupportsResources = v1.NewOptBool(plan.SupportsResources.ValueBool())
	createReq.Immutable = v1.NewOptBool(plan.Immutable.ValueBool())
	createReq.Active = v1.NewOptBool(plan.Active.ValueBool())

	if !plan.OAuthServiceID.IsNull() {
		oauthID, err := uuid.Parse(plan.OAuthServiceID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Invalid OAuth Service ID", err.Error())
			return
		}
		createReq.OAuthServiceID = v1.NewOptNilUUID(oauthID)
	}

	if len(allowedTools) > 0 {
		createReq.AllowedTools = v1.NewOptNilStringArray(allowedTools)
	}

	if len(deniedTools) > 0 {
		createReq.DeniedTools = v1.NewOptNilStringArray(deniedTools)
	}

	resultInterface, err := r.client.CreateMcpendpoint(ctx, &createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating MCP endpoint",
			"Could not create MCP endpoint: "+err.Error(),
		)
		return
	}

	// Type assert the response
	result, ok := resultInterface.(*v1.MCPEndpointResponse)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected response type",
			"Expected MCPEndpointResponse",
		)
		return
	}

	// Update state with created resource
	plan.ID = types.StringValue(result.ID.String())
	plan.Name = types.StringValue(result.Name)
	plan.URL = types.StringValue(result.URL)

	if result.Description.IsSet() {
		plan.Description = types.StringValue(result.Description.Value)
	}

	if result.Headers.IsSet() && len(result.Headers.Value) > 0 {
		headersMap := make(map[string]types.String)
		for k, v := range result.Headers.Value {
			headersMap[k] = types.StringValue(v)
		}
		plan.Headers = types.MapValueMust(types.StringType, convertMapToStringValues(headersMap))
	}

	if result.DevgraphAuth.IsSet() {
		plan.DevgraphAuth = types.BoolValue(result.DevgraphAuth.Value)
	}
	if result.SupportsResources.IsSet() {
		plan.SupportsResources = types.BoolValue(result.SupportsResources.Value)
	}
	if result.Immutable.IsSet() {
		plan.Immutable = types.BoolValue(result.Immutable.Value)
	}
	if result.Active.IsSet() {
		plan.Active = types.BoolValue(result.Active.Value)
	}

	if result.OAuthServiceID.IsSet() && result.OAuthServiceID.Value != uuid.Nil {
		plan.OAuthServiceID = types.StringValue(result.OAuthServiceID.Value.String())
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *MCPEndpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MCPEndpointResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpointID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid MCP Endpoint ID", err.Error())
		return
	}

	resultInterface, err := r.client.GetMcpendpoint(ctx, v1.GetMcpendpointParams{
		McpendpointID: endpointID,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading MCP endpoint",
			"Could not read MCP endpoint ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Check if resource was not found (deleted outside Terraform)
	if _, ok := resultInterface.(*v1.GetMcpendpointNotFound); ok {
		// Resource was deleted outside Terraform, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Type assert the response
	result, ok := resultInterface.(*v1.MCPEndpointResponse)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected response type",
			fmt.Sprintf("Expected *v1.MCPEndpointResponse, got: %T", resultInterface),
		)
		return
	}

	// Update state
	state.Name = types.StringValue(result.Name)
	state.URL = types.StringValue(result.URL)

	if result.Description.IsSet() {
		state.Description = types.StringValue(result.Description.Value)
	}

	if result.Headers.IsSet() && len(result.Headers.Value) > 0 {
		headersMap := make(map[string]types.String)
		for k, v := range result.Headers.Value {
			headersMap[k] = types.StringValue(v)
		}
		state.Headers = types.MapValueMust(types.StringType, convertMapToStringValues(headersMap))
	}

	if result.DevgraphAuth.IsSet() {
		state.DevgraphAuth = types.BoolValue(result.DevgraphAuth.Value)
	}
	if result.SupportsResources.IsSet() {
		state.SupportsResources = types.BoolValue(result.SupportsResources.Value)
	}
	if result.Immutable.IsSet() {
		state.Immutable = types.BoolValue(result.Immutable.Value)
	}
	if result.Active.IsSet() {
		state.Active = types.BoolValue(result.Active.Value)
	}

	if result.OAuthServiceID.IsSet() && result.OAuthServiceID.Value != uuid.Nil {
		state.OAuthServiceID = types.StringValue(result.OAuthServiceID.Value.String())
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *MCPEndpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MCPEndpointResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpointID, err := uuid.Parse(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid MCP Endpoint ID", err.Error())
		return
	}

	// Build update request
	updateReq := v1.MCPEndpointUpdate{}

	if !plan.Name.IsNull() {
		updateReq.Name = v1.NewOptNilString(plan.Name.ValueString())
	}

	if !plan.URL.IsNull() {
		updateReq.URL = v1.NewOptNilString(plan.URL.ValueString())
	}

	if !plan.Description.IsNull() {
		updateReq.Description = v1.NewOptNilString(plan.Description.ValueString())
	}

	if !plan.Headers.IsNull() {
		headers := make(map[string]string)
		diags = plan.Headers.ElementsAs(ctx, &headers, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		updateReq.Headers = v1.NewOptNilMCPEndpointUpdateHeaders(v1.MCPEndpointUpdateHeaders(headers))
	}

	if !plan.DevgraphAuth.IsNull() {
		updateReq.DevgraphAuth = v1.NewOptNilBool(plan.DevgraphAuth.ValueBool())
	}

	if !plan.SupportsResources.IsNull() {
		updateReq.SupportsResources = v1.NewOptNilBool(plan.SupportsResources.ValueBool())
	}

	if !plan.Immutable.IsNull() {
		updateReq.Immutable = v1.NewOptNilBool(plan.Immutable.ValueBool())
	}

	if !plan.Active.IsNull() {
		updateReq.Active = v1.NewOptNilBool(plan.Active.ValueBool())
	}

	if !plan.OAuthServiceID.IsNull() {
		oauthID, err := uuid.Parse(plan.OAuthServiceID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Invalid OAuth Service ID", err.Error())
			return
		}
		updateReq.OAuthServiceID = v1.NewOptNilUUID(oauthID)
	}

	if !plan.AllowedTools.IsNull() {
		var allowedTools []string
		diags = plan.AllowedTools.ElementsAs(ctx, &allowedTools, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		updateReq.AllowedTools = v1.NewOptNilStringArray(allowedTools)
	}

	if !plan.DeniedTools.IsNull() {
		var deniedTools []string
		diags = plan.DeniedTools.ElementsAs(ctx, &deniedTools, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		updateReq.DeniedTools = v1.NewOptNilStringArray(deniedTools)
	}

	resultInterface, err := r.client.UpdateMcpendpoint(ctx, &updateReq, v1.UpdateMcpendpointParams{
		McpendpointID: endpointID,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating MCP endpoint",
			"Could not update MCP endpoint: "+err.Error(),
		)
		return
	}

	// Type assert the response
	result, ok := resultInterface.(*v1.MCPEndpointResponse)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected response type",
			"Expected MCPEndpointResponse",
		)
		return
	}

	// Update state
	plan.Name = types.StringValue(result.Name)
	plan.URL = types.StringValue(result.URL)

	if result.Description.IsSet() {
		plan.Description = types.StringValue(result.Description.Value)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *MCPEndpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MCPEndpointResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpointID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid MCP Endpoint ID", err.Error())
		return
	}

	_, err = r.client.DeleteMcpendpoint(ctx, v1.DeleteMcpendpointParams{
		McpendpointID: endpointID,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting MCP endpoint",
			"Could not delete MCP endpoint: "+err.Error(),
		)
		return
	}
}

func (r *MCPEndpointResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper function to convert map[string]types.String to map[string]attr.Value
func convertMapToStringValues(m map[string]types.String) map[string]attr.Value {
	result := make(map[string]attr.Value, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}
