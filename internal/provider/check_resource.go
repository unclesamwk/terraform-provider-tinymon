package provider

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &checkResource{}
	_ resource.ResourceWithImportState = &checkResource{}
)

func NewCheckResource() resource.Resource {
	return &checkResource{}
}

type checkResource struct {
	client *TinyMonClient
}

type checkResourceModel struct {
	ID              types.Int64  `tfsdk:"id"`
	HostAddress     types.String `tfsdk:"host_address"`
	Type            types.String `tfsdk:"type"`
	Config          types.String `tfsdk:"config"`
	IntervalSeconds types.Int64  `tfsdk:"interval_seconds"`
	Enabled         types.Bool   `tfsdk:"enabled"`
}

type checkAPIRequest struct {
	HostAddress     string `json:"host_address"`
	Type            string `json:"type"`
	Config          string `json:"config"`
	IntervalSeconds int64  `json:"interval_seconds"`
	Enabled         int    `json:"enabled"`
}

type checkAPIResponse struct {
	ID              int64  `json:"id"`
	HostID          int64  `json:"host_id"`
	Type            string `json:"type"`
	Config          string `json:"config"`
	IntervalSeconds int64  `json:"interval_seconds"`
	Enabled         int    `json:"enabled"`
}

type checkDeleteRequest struct {
	HostAddress string `json:"host_address"`
	Type        string `json:"type"`
	Config      string `json:"config"`
}

func (r *checkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_check"
}

func (r *checkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a check in TinyMon.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"host_address": schema.StringAttribute{
				Description: "Address of the host. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "Check type (ping, http, port, certificate, etc.).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"config": schema.StringAttribute{
				Description: "JSON config string.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("{}"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"interval_seconds": schema.Int64Attribute{
				Description: "Check interval in seconds.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(300),
			},
			"enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
		},
	}
}

func (r *checkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *checkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan checkResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	enabled := 1
	if !plan.Enabled.IsNull() && !plan.Enabled.ValueBool() {
		enabled = 0
	}

	body := checkAPIRequest{
		HostAddress:     plan.HostAddress.ValueString(),
		Type:            plan.Type.ValueString(),
		Config:          plan.Config.ValueString(),
		IntervalSeconds: plan.IntervalSeconds.ValueInt64(),
		Enabled:         enabled,
	}

	var result checkAPIResponse
	if err := r.client.DoJSON("POST", "/api/push/checks", body, &result); err != nil {
		resp.Diagnostics.AddError("Error creating check", err.Error())
		return
	}

	mapCheckResponseToState(&result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *checkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state checkResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	queryPath := fmt.Sprintf("/api/push/checks?host_address=%s&type=%s&config=%s",
		url.QueryEscape(state.HostAddress.ValueString()),
		url.QueryEscape(state.Type.ValueString()),
		url.QueryEscape(state.Config.ValueString()),
	)

	var result checkAPIResponse
	if err := r.client.DoJSON("GET", queryPath, nil, &result); err != nil {
		resp.Diagnostics.AddError("Error reading check", err.Error())
		return
	}

	mapCheckResponseToState(&result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *checkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan checkResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	enabled := 1
	if !plan.Enabled.IsNull() && !plan.Enabled.ValueBool() {
		enabled = 0
	}

	body := checkAPIRequest{
		HostAddress:     plan.HostAddress.ValueString(),
		Type:            plan.Type.ValueString(),
		Config:          plan.Config.ValueString(),
		IntervalSeconds: plan.IntervalSeconds.ValueInt64(),
		Enabled:         enabled,
	}

	var result checkAPIResponse
	if err := r.client.DoJSON("POST", "/api/push/checks", body, &result); err != nil {
		resp.Diagnostics.AddError("Error updating check", err.Error())
		return
	}

	mapCheckResponseToState(&result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *checkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state checkResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := checkDeleteRequest{
		HostAddress: state.HostAddress.ValueString(),
		Type:        state.Type.ValueString(),
		Config:      state.Config.ValueString(),
	}

	if err := r.client.DoJSON("DELETE", "/api/push/checks", body, nil); err != nil {
		resp.Diagnostics.AddError("Error deleting check", err.Error())
		return
	}
}

func (r *checkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) < 2 {
		resp.Diagnostics.AddError("Invalid import ID",
			"Import ID must be: host_address/type or host_address/type/config")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("host_address"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), parts[1])...)

	config := "{}"
	if len(parts) == 3 {
		config = parts[2]
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("config"), config)...)
}

func mapCheckResponseToState(apiResp *checkAPIResponse, state *checkResourceModel) {
	state.ID = types.Int64Value(apiResp.ID)
	state.Type = types.StringValue(apiResp.Type)
	state.Config = types.StringValue(apiResp.Config)
	state.IntervalSeconds = types.Int64Value(apiResp.IntervalSeconds)
	state.Enabled = types.BoolValue(apiResp.Enabled != 0)
}
