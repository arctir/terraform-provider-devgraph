package provider

import (
	"context"
	"fmt"

	v1 "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &ModelResource{}
	_ resource.ResourceWithConfigure   = &ModelResource{}
	_ resource.ResourceWithImportState = &ModelResource{}
)

func NewModelResource() resource.Resource {
	return &ModelResource{}
}

type ModelResource struct {
	client *v1.Client
}

type ModelResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	ProviderID  types.String `tfsdk:"provider_id"`
	Default     types.Bool   `tfsdk:"default"`
}

func (r *ModelResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_model"
}

func (r *ModelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Model configuration in Devgraph.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the model.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the model (e.g., 'gpt-4', 'claude-3-opus').",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the model.",
				Optional:    true,
			},
			"provider_id": schema.StringAttribute{
				Description: "The ID of the model provider this model belongs to.",
				Required:    true,
			},
			"default": schema.BoolAttribute{
				Description: "Whether this is the default model.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (r *ModelResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ModelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ModelResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	providerID, err := uuid.Parse(plan.ProviderID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Provider ID", err.Error())
		return
	}

	createReq := v1.ModelCreate{
		Name:       plan.Name.ValueString(),
		ProviderID: providerID,
		Default:    v1.NewOptBool(plan.Default.ValueBool()),
	}

	if !plan.Description.IsNull() {
		createReq.Description = v1.NewOptNilString(plan.Description.ValueString())
	}

	resultInterface, err := r.client.CreateModel(ctx, &createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating model",
			"Could not create model: "+err.Error(),
		)
		return
	}

	// Type assert the response
	result, ok := resultInterface.(*v1.ModelResponse)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected response type",
			"Expected ModelResponse",
		)
		return
	}

	// Update state with created resource
	plan.ID = types.StringValue(result.ID.String())
	plan.Name = types.StringValue(result.Name)
	if result.Description.IsSet() {
		plan.Description = types.StringValue(result.Description.Value)
	}
	plan.ProviderID = types.StringValue(result.ProviderID.String())
	if result.Default.IsSet() {
		plan.Default = types.BoolValue(result.Default.Value)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ModelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ModelResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resultInterface, err := r.client.GetModel(ctx, v1.GetModelParams{
		ModelName: state.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading model",
			"Could not read model "+state.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	// Check if resource was not found (deleted outside Terraform)
	if _, ok := resultInterface.(*v1.GetModelNotFound); ok {
		// Resource was deleted outside Terraform, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Type assert the response
	result, ok := resultInterface.(*v1.ModelResponse)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected response type",
			fmt.Sprintf("Expected *v1.ModelResponse, got: %T", resultInterface),
		)
		return
	}

	// Update state
	state.ID = types.StringValue(result.ID.String())
	state.Name = types.StringValue(result.Name)
	if result.Description.IsSet() {
		state.Description = types.StringValue(result.Description.Value)
	}
	state.ProviderID = types.StringValue(result.ProviderID.String())
	if result.Default.IsSet() {
		state.Default = types.BoolValue(result.Default.Value)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *ModelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ModelResourceModel
	var state ModelResourceModel

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

	// Build update request
	updateReq := v1.ModelUpdate{}

	// Only include fields that have changed
	if !plan.Description.Equal(state.Description) {
		if plan.Description.IsNull() {
			updateReq.Description = v1.NewOptNilString("")
		} else {
			updateReq.Description = v1.NewOptNilString(plan.Description.ValueString())
		}
	}
	if !plan.ProviderID.Equal(state.ProviderID) {
		providerID, err := uuid.Parse(plan.ProviderID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Invalid Provider ID", err.Error())
			return
		}
		updateReq.ProviderID = v1.NewOptNilUUID(providerID)
	}
	if !plan.Default.Equal(state.Default) {
		updateReq.Default = v1.NewOptNilBool(plan.Default.ValueBool())
	}

	resultInterface, err := r.client.UpdateModel(ctx, &updateReq, v1.UpdateModelParams{
		ModelName: state.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating model",
			"Could not update model: "+err.Error(),
		)
		return
	}

	// Type assert the response
	result, ok := resultInterface.(*v1.ModelResponse)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected response type",
			"Expected ModelResponse",
		)
		return
	}

	// Update state with updated resource
	plan.ID = types.StringValue(result.ID.String())
	plan.Name = types.StringValue(result.Name)
	if result.Description.IsSet() {
		plan.Description = types.StringValue(result.Description.Value)
	}
	plan.ProviderID = types.StringValue(result.ProviderID.String())
	if result.Default.IsSet() {
		plan.Default = types.BoolValue(result.Default.Value)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ModelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ModelResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteModel(ctx, v1.DeleteModelParams{
		ModelName: state.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting model",
			"Could not delete model: "+err.Error(),
		)
		return
	}
}

func (r *ModelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// For models, we import by name, not ID
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
