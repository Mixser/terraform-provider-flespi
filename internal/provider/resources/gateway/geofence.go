package gateway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mixser/flespi-client"
	flespi_geofence "github.com/mixser/flespi-client/resources/geofence"
)

var (
	_ resource.Resource              = &gwDeviceResource{}
	_ resource.ResourceWithConfigure = &gwDeviceResource{}
)

type gwGeofenceResource struct {
	client *flespi.Client
}

type geofenceResourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`

	Enabled  types.Bool  `tfsdk:"enabled"`
	Priority types.Int64 `tfsdk:"priority"`

	Geometry jsontypes.NormalizedType `tfsdk:"geometry"`
}

func NewGeofenceResource() resource.Resource {
	return &gwGeofenceResource{}
}

func (g *gwGeofenceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_geofence"
}

func (g *gwGeofenceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"priority": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"geometry": jsontypes.Normalized{
				Required: true,
			},
		},
	}
}

func (g *gwGeofenceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*flespi.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *flespi.Client, got: %T", req.ProviderData),
		)
		return
	}

	g.client = client
}

func (g *gwGeofenceResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data *geofenceResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

}

func (g *gwGeofenceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data *geofenceResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

}

func (g *gwGeofenceResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data *geofenceResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

}

func (g *gwGeofenceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data *geofenceResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}
}

func (g *gwGeofenceResource) convertResourceModelToFlespiGeofence(ctx context.Context, data geofenceResourceModel) (flespi_geofence.Geofence, diag.Diagnostic) {

}

func (g *gwGeofenceResource) convertFlespiGeofenceToResourceModel(ctx context.Context, data flespi_geofence.Geofence) (geofenceResourceModel, diag.Diagnostic) {

}
