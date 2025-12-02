package provider

import (
	"context"
	"fmt"

	v1 "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &EnvironmentResource{}
	_ resource.ResourceWithConfigure   = &EnvironmentResource{}
	_ resource.ResourceWithImportState = &EnvironmentResource{}
)

func NewEnvironmentResource() resource.Resource {
	return &EnvironmentResource{}
}

type EnvironmentResource struct {
	client *v1.Client
}

type EnvironmentResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Slug                 types.String `tfsdk:"slug"`
	ClerkOrganizationID  types.String `tfsdk:"clerk_organization_id"`
	CustomerID           types.String `tfsdk:"customer_id"`
	SubscriptionID       types.String `tfsdk:"subscription_id"`
	InvitedUsers         types.List   `tfsdk:"invited_users"`
	StripeSubscriptionID types.String `tfsdk:"stripe_subscription_id"`
	InstanceURL          types.String `tfsdk:"instance_url"`
}

func (r *EnvironmentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (r *EnvironmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an environment in Devgraph.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the environment.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the environment.",
				Required:    true,
			},
			"slug": schema.StringAttribute{
				Description: "The URL-friendly slug of the environment.",
				Computed:    true,
			},
			"clerk_organization_id": schema.StringAttribute{
				Description: "The Clerk organization ID associated with this environment.",
				Computed:    true,
			},
			"customer_id": schema.StringAttribute{
				Description: "The customer ID associated with this environment.",
				Computed:    true,
			},
			"subscription_id": schema.StringAttribute{
				Description: "The subscription ID associated with this environment.",
				Computed:    true,
			},
			"invited_users": schema.ListAttribute{
				Description: "List of email addresses to invite to this environment.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"stripe_subscription_id": schema.StringAttribute{
				Description: "The Stripe subscription ID for this environment.",
				Required:    true,
			},
			"instance_url": schema.StringAttribute{
				Description: "The instance URL for this environment.",
				Required:    true,
			},
		},
	}
}

func (r *EnvironmentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EnvironmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EnvironmentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build invited users list
	var invitedUsers []string
	if !plan.InvitedUsers.IsNull() {
		diags = plan.InvitedUsers.ElementsAs(ctx, &invitedUsers, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Create environment
	createReq := v1.EnvironmentCreate{
		Name:                 plan.Name.ValueString(),
		StripeSubscriptionID: plan.StripeSubscriptionID.ValueString(),
		InstanceURL:          plan.InstanceURL.ValueString(),
		InvitedUsers:         invitedUsers,
	}

	res, err := r.client.CreateEnvironment(ctx, &createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating environment",
			"Could not create environment: "+err.Error(),
		)
		return
	}

	// Type assert the response
	result, ok := res.(*v1.EnvironmentResponse)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected response type",
			fmt.Sprintf("Expected *v1.EnvironmentResponse, got: %T", res),
		)
		return
	}

	// Update state with created resource
	plan.ID = types.StringValue(result.ID.String())
	plan.Name = types.StringValue(result.Name)
	plan.Slug = types.StringValue(result.Slug)
	plan.ClerkOrganizationID = types.StringValue(result.ClerkOrganizationID)
	plan.CustomerID = types.StringValue(result.CustomerID)
	plan.SubscriptionID = types.StringValue(result.SubscriptionID.String())

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *EnvironmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EnvironmentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Note: The API spec shows get_environments returns a list, not a single environment
	// We'll need to filter by ID from the list
	res, err := r.client.GetEnvironments(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading environments",
			"Could not read environments: "+err.Error(),
		)
		return
	}

	// Type assert the response
	okResponse, ok := res.(*v1.GetEnvironmentsOKApplicationJSON)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected response type",
			fmt.Sprintf("Expected *v1.GetEnvironmentsOKApplicationJSON, got: %T", res),
		)
		return
	}

	environments := []v1.EnvironmentResponse(*okResponse)

	// Find the environment with matching ID
	var found bool
	for _, env := range environments {
		if env.ID.String() == state.ID.ValueString() {
			state.Name = types.StringValue(env.Name)
			state.Slug = types.StringValue(env.Slug)
			state.ClerkOrganizationID = types.StringValue(env.ClerkOrganizationID)
			state.CustomerID = types.StringValue(env.CustomerID)
			state.SubscriptionID = types.StringValue(env.SubscriptionID.String())
			found = true
			break
		}
	}

	if !found {
		// Resource was deleted outside Terraform, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *EnvironmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// According to the API spec, there's no update endpoint for environments
	// This is a placeholder that will return an error if an update is attempted
	resp.Diagnostics.AddError(
		"Update not supported",
		"Environments cannot be updated after creation. Please destroy and recreate the resource.",
	)
}

func (r *EnvironmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// According to the API spec, there's no delete endpoint for environments
	// This is a placeholder that will return an error if deletion is attempted
	resp.Diagnostics.AddError(
		"Delete not supported",
		"Environments cannot be deleted through the API. Please manage environment deletion through the Devgraph console.",
	)
}

func (r *EnvironmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
