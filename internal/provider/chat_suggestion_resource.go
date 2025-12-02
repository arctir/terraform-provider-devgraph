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
	_ resource.Resource                = &ChatSuggestionResource{}
	_ resource.ResourceWithConfigure   = &ChatSuggestionResource{}
	_ resource.ResourceWithImportState = &ChatSuggestionResource{}
)

func NewChatSuggestionResource() resource.Resource {
	return &ChatSuggestionResource{}
}

type ChatSuggestionResource struct {
	client *v1.Client
}

type ChatSuggestionResourceModel struct {
	ID     types.String `tfsdk:"id"`
	Title  types.String `tfsdk:"title"`
	Label  types.String `tfsdk:"label"`
	Action types.String `tfsdk:"action"`
	Active types.Bool   `tfsdk:"active"`
}

func (r *ChatSuggestionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_chat_suggestion"
}

func (r *ChatSuggestionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a chat suggestion in Devgraph. Chat suggestions are quick-start prompts shown to users in the chat interface.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the chat suggestion.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				Description: "The title of the suggestion displayed to users.",
				Required:    true,
			},
			"label": schema.StringAttribute{
				Description: "A short label or category for the suggestion.",
				Required:    true,
			},
			"action": schema.StringAttribute{
				Description: "The action or prompt text that will be used when the suggestion is clicked.",
				Required:    true,
			},
			"active": schema.BoolAttribute{
				Description: "Whether this suggestion is active and should be shown to users.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
		},
	}
}

func (r *ChatSuggestionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ChatSuggestionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ChatSuggestionResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build create request
	createReq := v1.ChatSuggestionCreate{
		Title:  plan.Title.ValueString(),
		Label:  plan.Label.ValueString(),
		Action: plan.Action.ValueString(),
	}

	if !plan.Active.IsNull() {
		active := plan.Active.ValueBool()
		createReq.SetActive(v1.NewOptBool(active))
	}

	// Create chat suggestion
	res, err := r.client.CreateChatSuggestion(ctx, &createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating chat suggestion",
			"Could not create chat suggestion: "+err.Error(),
		)
		return
	}

	// Type assert the response
	result, ok := res.(*v1.ChatSuggestionResponse)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected response type",
			fmt.Sprintf("Expected *v1.ChatSuggestionResponse, got: %T", res),
		)
		return
	}

	// Update state with created resource
	plan.ID = types.StringValue(result.ID.String())
	plan.Title = types.StringValue(result.Title)
	plan.Label = types.StringValue(result.Label)
	plan.Action = types.StringValue(result.Action)
	if result.Active.IsSet() {
		plan.Active = types.BoolValue(result.Active.Value)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ChatSuggestionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ChatSuggestionResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse UUID
	suggestionID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid suggestion ID",
			"Could not parse suggestion ID as UUID: "+err.Error(),
		)
		return
	}

	// List all suggestions and find the one with matching ID
	// Note: The API doesn't have a GetChatSuggestion endpoint, only List
	res, err := r.client.ListChatSuggestions(ctx, v1.ListChatSuggestionsParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading chat suggestions",
			"Could not read chat suggestions: "+err.Error(),
		)
		return
	}

	// Type assert the response
	listResult, ok := res.(*v1.ListChatSuggestionsOKApplicationJSON)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected response type",
			fmt.Sprintf("Expected *v1.ListChatSuggestionsOKApplicationJSON, got: %T", res),
		)
		return
	}

	suggestions := []v1.ChatSuggestionResponse(*listResult)

	// Find the suggestion with matching ID
	var found bool
	for _, suggestion := range suggestions {
		if suggestion.ID == suggestionID {
			state.Title = types.StringValue(suggestion.Title)
			state.Label = types.StringValue(suggestion.Label)
			state.Action = types.StringValue(suggestion.Action)
			if suggestion.Active.IsSet() {
				state.Active = types.BoolValue(suggestion.Active.Value)
			}
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

func (r *ChatSuggestionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// According to the API, there's no update endpoint for chat suggestions
	// To update, we need to delete and recreate
	resp.Diagnostics.AddError(
		"Update not supported",
		"Chat suggestions cannot be updated. Please destroy and recreate the resource to make changes.",
	)
}

func (r *ChatSuggestionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ChatSuggestionResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse UUID
	suggestionID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid suggestion ID",
			"Could not parse suggestion ID as UUID: "+err.Error(),
		)
		return
	}

	// Delete chat suggestion
	_, err = r.client.DeleteChatSuggestion(ctx, v1.DeleteChatSuggestionParams{
		SuggestionID: suggestionID,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting chat suggestion",
			"Could not delete chat suggestion: "+err.Error(),
		)
		return
	}
}

func (r *ChatSuggestionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
