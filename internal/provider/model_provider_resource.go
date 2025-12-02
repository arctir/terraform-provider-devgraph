package provider

import (
	"context"
	"fmt"

	v1 "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &ModelProviderResource{}
	_ resource.ResourceWithConfigure   = &ModelProviderResource{}
	_ resource.ResourceWithImportState = &ModelProviderResource{}
)

func NewModelProviderResource() resource.Resource {
	return &ModelProviderResource{}
}

type ModelProviderResource struct {
	client *v1.Client
}

type ModelProviderResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Type    types.String `tfsdk:"type"`
	Name    types.String `tfsdk:"name"`
	APIKey  types.String `tfsdk:"api_key"`
	Default types.Bool   `tfsdk:"default"`
}

func (r *ModelProviderResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_model_provider"
}

func (r *ModelProviderResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Model Provider configuration in Devgraph.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the model provider.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Description: "The type of model provider (openai, anthropic, xai).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("openai", "anthropic", "xai"),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the model provider.",
				Required:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "The API key for the model provider.",
				Required:    true,
				Sensitive:   true,
			},
			"default": schema.BoolAttribute{
				Description: "Whether this is the default model provider.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (r *ModelProviderResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ModelProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ModelProviderResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the appropriate provider type based on the type field
	var createReq v1.ModelProviderCreate
	providerType := plan.Type.ValueString()

	switch providerType {
	case "openai":
		createReq = v1.ModelProviderCreate{
			Data: v1.ModelProviderCreateData{
				Type: v1.OpenAIModelProviderCreateModelProviderCreateData,
				OpenAIModelProviderCreate: v1.OpenAIModelProviderCreate{
					Type:    "openai",
					Name:    plan.Name.ValueString(),
					APIKey:  plan.APIKey.ValueString(),
					Default: v1.NewOptBool(plan.Default.ValueBool()),
				},
			},
		}
	case "anthropic":
		createReq = v1.ModelProviderCreate{
			Data: v1.ModelProviderCreateData{
				Type: v1.AnthropicModelProviderCreateModelProviderCreateData,
				AnthropicModelProviderCreate: v1.AnthropicModelProviderCreate{
					Type:    "anthropic",
					Name:    plan.Name.ValueString(),
					APIKey:  plan.APIKey.ValueString(),
					Default: v1.NewOptBool(plan.Default.ValueBool()),
				},
			},
		}
	case "xai":
		createReq = v1.ModelProviderCreate{
			Data: v1.ModelProviderCreateData{
				Type: v1.XAIModelProviderCreateModelProviderCreateData,
				XAIModelProviderCreate: v1.XAIModelProviderCreate{
					Type:    "xai",
					Name:    plan.Name.ValueString(),
					APIKey:  plan.APIKey.ValueString(),
					Default: v1.NewOptBool(plan.Default.ValueBool()),
				},
			},
		}
	default:
		resp.Diagnostics.AddError(
			"Invalid provider type",
			fmt.Sprintf("Provider type must be one of: openai, anthropic, xai. Got: %s", providerType),
		)
		return
	}

	resultInterface, err := r.client.CreateModelprovider(ctx, &createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating model provider",
			"Could not create model provider: "+err.Error(),
		)
		return
	}

	// Type assert the response
	result, ok := resultInterface.(*v1.ModelProviderResponse)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected response type",
			"Expected ModelProviderResponse",
		)
		return
	}

	// Update state with created resource based on type
	switch result.Type {
	case v1.OpenAIModelProviderResponseModelProviderResponse:
		provider := result.OpenAIModelProviderResponse
		plan.ID = types.StringValue(provider.ID.String())
		plan.Type = types.StringValue(provider.Type)
		plan.Name = types.StringValue(provider.Name)
		plan.APIKey = types.StringValue(provider.APIKey)
		if provider.Default.IsSet() {
			plan.Default = types.BoolValue(provider.Default.Value)
		}
	case v1.AnthropicModelProviderResponseModelProviderResponse:
		provider := result.AnthropicModelProviderResponse
		plan.ID = types.StringValue(provider.ID.String())
		plan.Type = types.StringValue(provider.Type)
		plan.Name = types.StringValue(provider.Name)
		plan.APIKey = types.StringValue(provider.APIKey)
		if provider.Default.IsSet() {
			plan.Default = types.BoolValue(provider.Default.Value)
		}
	case v1.XAIModelProviderResponseModelProviderResponse:
		provider := result.XAIModelProviderResponse
		plan.ID = types.StringValue(provider.ID.String())
		plan.Type = types.StringValue(provider.Type)
		plan.Name = types.StringValue(provider.Name)
		plan.APIKey = types.StringValue(provider.APIKey)
		if provider.Default.IsSet() {
			plan.Default = types.BoolValue(provider.Default.Value)
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ModelProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ModelProviderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	providerID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Model Provider ID", err.Error())
		return
	}

	resultInterface, err := r.client.GetModelprovider(ctx, v1.GetModelproviderParams{
		ProviderID: providerID,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading model provider",
			"Could not read model provider ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Check if resource was not found (deleted outside Terraform)
	if _, ok := resultInterface.(*v1.GetModelproviderNotFound); ok {
		// Resource was deleted outside Terraform, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Type assert the response
	result, ok := resultInterface.(*v1.ModelProviderResponse)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected response type",
			fmt.Sprintf("Expected *v1.ModelProviderResponse, got: %T", resultInterface),
		)
		return
	}

	// Update state based on provider type
	switch result.Type {
	case v1.OpenAIModelProviderResponseModelProviderResponse:
		provider := result.OpenAIModelProviderResponse
		state.Type = types.StringValue(provider.Type)
		state.Name = types.StringValue(provider.Name)
		state.APIKey = types.StringValue(provider.APIKey)
		if provider.Default.IsSet() {
			state.Default = types.BoolValue(provider.Default.Value)
		}
	case v1.AnthropicModelProviderResponseModelProviderResponse:
		provider := result.AnthropicModelProviderResponse
		state.Type = types.StringValue(provider.Type)
		state.Name = types.StringValue(provider.Name)
		state.APIKey = types.StringValue(provider.APIKey)
		if provider.Default.IsSet() {
			state.Default = types.BoolValue(provider.Default.Value)
		}
	case v1.XAIModelProviderResponseModelProviderResponse:
		provider := result.XAIModelProviderResponse
		state.Type = types.StringValue(provider.Type)
		state.Name = types.StringValue(provider.Name)
		state.APIKey = types.StringValue(provider.APIKey)
		if provider.Default.IsSet() {
			state.Default = types.BoolValue(provider.Default.Value)
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *ModelProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ModelProviderResourceModel
	var state ModelProviderResourceModel

	// Get plan and state
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	providerID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Model Provider ID", err.Error())
		return
	}

	// Build update request
	updateReq := v1.ModelProviderUpdate{}

	// Only include fields that have changed
	if !plan.Name.Equal(state.Name) {
		updateReq.Name = v1.NewOptNilString(plan.Name.ValueString())
	}
	if !plan.APIKey.Equal(state.APIKey) {
		updateReq.APIKey = v1.NewOptNilString(plan.APIKey.ValueString())
	}
	if !plan.Default.Equal(state.Default) {
		updateReq.Default = v1.NewOptNilBool(plan.Default.ValueBool())
	}

	resultInterface, err := r.client.UpdateModelprovider(ctx, &updateReq, v1.UpdateModelproviderParams{
		ProviderID: providerID,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating model provider",
			"Could not update model provider: "+err.Error(),
		)
		return
	}

	// Type assert the response
	result, ok := resultInterface.(*v1.ModelProviderResponse)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected response type",
			"Expected ModelProviderResponse",
		)
		return
	}

	// Update state with updated resource based on type
	switch result.Type {
	case v1.OpenAIModelProviderResponseModelProviderResponse:
		provider := result.OpenAIModelProviderResponse
		plan.ID = types.StringValue(provider.ID.String())
		plan.Type = types.StringValue(provider.Type)
		plan.Name = types.StringValue(provider.Name)
		plan.APIKey = types.StringValue(provider.APIKey)
		if provider.Default.IsSet() {
			plan.Default = types.BoolValue(provider.Default.Value)
		}
	case v1.AnthropicModelProviderResponseModelProviderResponse:
		provider := result.AnthropicModelProviderResponse
		plan.ID = types.StringValue(provider.ID.String())
		plan.Type = types.StringValue(provider.Type)
		plan.Name = types.StringValue(provider.Name)
		plan.APIKey = types.StringValue(provider.APIKey)
		if provider.Default.IsSet() {
			plan.Default = types.BoolValue(provider.Default.Value)
		}
	case v1.XAIModelProviderResponseModelProviderResponse:
		provider := result.XAIModelProviderResponse
		plan.ID = types.StringValue(provider.ID.String())
		plan.Type = types.StringValue(provider.Type)
		plan.Name = types.StringValue(provider.Name)
		plan.APIKey = types.StringValue(provider.APIKey)
		if provider.Default.IsSet() {
			plan.Default = types.BoolValue(provider.Default.Value)
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ModelProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ModelProviderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	providerID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Model Provider ID", err.Error())
		return
	}

	_, err = r.client.DeleteModelprovider(ctx, v1.DeleteModelproviderParams{
		ProviderID: providerID,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting model provider",
			"Could not delete model provider: "+err.Error(),
		)
		return
	}
}

func (r *ModelProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
