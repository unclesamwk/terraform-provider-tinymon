package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &hostResource{}
	_ resource.ResourceWithImportState = &hostResource{}
)

func NewHostResource() resource.Resource {
	return &hostResource{}
}

type hostResource struct {
	client *TinyMonClient
}

type hostResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Address     types.String `tfsdk:"address"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Topic       types.String `tfsdk:"topic"`
	Enabled     types.Bool   `tfsdk:"enabled"`
}

type hostAPIRequest struct {
	Address     string `json:"address"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description"`
	Topic       string `json:"topic"`
	Enabled     int    `json:"enabled"`
}

type hostAPIResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	Description string `json:"description"`
	Topic       string `json:"topic"`
	Enabled     int    `json:"enabled"`
}

type hostDeleteRequest struct {
	Address string `json:"address"`
}

func (r *hostResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host"
}

func (r *hostResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a host in TinyMon.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"address": schema.StringAttribute{
				Description: "IP address or hostname. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Display name. Defaults to the address.",
				Optional:    true,
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"topic": schema.StringAttribute{
				Description: "Topic path for grouping.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
		},
	}
}

func (r *hostResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*TinyMonClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *TinyMonClient, got %T.", req.ProviderData))
		return
	}
	r.client = client
}

func (r *hostResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan hostResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	enabled := 1
	if !plan.Enabled.IsNull() && !plan.Enabled.ValueBool() {
		enabled = 0
	}

	body := hostAPIRequest{
		Address:     plan.Address.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Topic:       plan.Topic.ValueString(),
		Enabled:     enabled,
	}

	var result hostAPIResponse
	if err := r.client.DoJSON("POST", "/api/push/hosts", body, &result); err != nil {
		resp.Diagnostics.AddError("Error creating host", err.Error())
		return
	}

	mapHostResponseToState(&result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *hostResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state hostResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiPath := "/api/push/hosts?address=" + url.QueryEscape(state.Address.ValueString())

	var result hostAPIResponse
	if err := r.client.DoJSON("GET", apiPath, nil, &result); err != nil {
		resp.Diagnostics.AddError("Error reading host", err.Error())
		return
	}

	mapHostResponseToState(&result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *hostResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan hostResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	enabled := 1
	if !plan.Enabled.IsNull() && !plan.Enabled.ValueBool() {
		enabled = 0
	}

	body := hostAPIRequest{
		Address:     plan.Address.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Topic:       plan.Topic.ValueString(),
		Enabled:     enabled,
	}

	var result hostAPIResponse
	if err := r.client.DoJSON("POST", "/api/push/hosts", body, &result); err != nil {
		resp.Diagnostics.AddError("Error updating host", err.Error())
		return
	}

	mapHostResponseToState(&result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *hostResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state hostResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := hostDeleteRequest{Address: state.Address.ValueString()}
	if err := r.client.DoJSON("DELETE", "/api/push/hosts", body, nil); err != nil {
		resp.Diagnostics.AddError("Error deleting host", err.Error())
		return
	}
}

func (r *hostResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("address"), req.ID)...)
}

func mapHostResponseToState(apiResp *hostAPIResponse, state *hostResourceModel) {
	state.ID = types.Int64Value(apiResp.ID)
	state.Address = types.StringValue(apiResp.Address)
	state.Name = types.StringValue(apiResp.Name)
	state.Description = types.StringValue(apiResp.Description)
	state.Topic = types.StringValue(apiResp.Topic)
	state.Enabled = types.BoolValue(apiResp.Enabled != 0)
}
