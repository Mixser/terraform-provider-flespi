package gateway

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mixser/flespi-client"
	flespi_geofence "github.com/mixser/flespi-client/resources/gateway/geofence"
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

	Geometry jsontypes.Normalized `tfsdk:"geometry"`
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
			"geometry": schema.StringAttribute{
				CustomType:  jsontypes.NormalizedType{},
				Required:    true,
				Description: "GeoJSON geometry (circle, polygon, or corridor)",
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

	instance, diags := g.convertResourceModelToFlespiGeofence(*data)

	if diags != nil {
		response.Diagnostics.Append(diags)
		return
	}

	geofenceInstance, err := flespi_geofence.NewGeofence(
		g.client,
		instance.Name,
		flespi_geofence.WithStatus(instance.Enabled),
		flespi_geofence.WithPriority(instance.Priority),
		flespi_geofence.WithGeometry(instance.Geometry),
	)

	if err != nil {
		response.Diagnostics.AddError(
			"Failed to create geofence",
			fmt.Sprintf("Error creating geofence: %s", err),
		)
		return
	}

	result, diags := g.convertFlespiGeofenceToResourceModel(*geofenceInstance)

	response.Diagnostics.Append(diags)
	response.Diagnostics.Append(response.State.Set(ctx, &result)...)
}

func (g *gwGeofenceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data *geofenceResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	geofences, err := flespi_geofence.ListGeofences(g.client)

	if err != nil {
		response.Diagnostics.AddError(
			"Failed to read geofences",
			fmt.Sprintf("Error reading geofences: %s", err),
		)
		return
	}

	// Find the geofence by ID
	var foundGeofence *flespi_geofence.Geofence
	for _, gf := range geofences {
		if gf.Id == data.ID.ValueInt64() {
			foundGeofence = &gf
			break
		}
	}

	if foundGeofence == nil {
		response.Diagnostics.AddError(
			"Geofence not found",
			fmt.Sprintf("Geofence with ID %d not found", data.ID.ValueInt64()),
		)
		return
	}

	result, diags := g.convertFlespiGeofenceToResourceModel(*foundGeofence)

	response.Diagnostics.Append(diags)
	response.Diagnostics.Append(response.State.Set(ctx, &result)...)
}

func (g *gwGeofenceResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data *geofenceResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	instance, diags := g.convertResourceModelToFlespiGeofence(*data)

	if diags != nil {
		response.Diagnostics.Append(diags)
		return
	}

	_, err := flespi_geofence.UpdateGeofence(g.client, instance)

	if err != nil {
		response.Diagnostics.AddError(
			"Failed to update geofence",
			fmt.Sprintf("Error updating geofence: %s", err),
		)
		return
	}

	// Re-read to get updated state
	geofences, err := flespi_geofence.ListGeofences(g.client)

	if err != nil {
		response.Diagnostics.AddError(
			"Failed to read geofence after update",
			fmt.Sprintf("Error reading geofence: %s", err),
		)
		return
	}

	var foundGeofence *flespi_geofence.Geofence
	for _, gf := range geofences {
		if gf.Id == data.ID.ValueInt64() {
			foundGeofence = &gf
			break
		}
	}

	if foundGeofence == nil {
		response.Diagnostics.AddError(
			"Geofence not found after update",
			fmt.Sprintf("Geofence with ID %d not found", data.ID.ValueInt64()),
		)
		return
	}

	result, diags := g.convertFlespiGeofenceToResourceModel(*foundGeofence)

	response.Diagnostics.Append(diags)
	response.Diagnostics.Append(response.State.Set(ctx, &result)...)
}

func (g *gwGeofenceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data *geofenceResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	err := flespi_geofence.DeleteGeofenceById(g.client, data.ID.ValueInt64())

	if err != nil {
		response.Diagnostics.AddError(
			"Failed to delete geofence",
			fmt.Sprintf("Error deleting geofence: %s", err),
		)
		return
	}
}

func (g *gwGeofenceResource) convertResourceModelToFlespiGeofence(data geofenceResourceModel) (flespi_geofence.Geofence, diag.Diagnostic) {
	var geometry flespi_geofence.GeofenceGeometry

	// Parse geometry from JSON
	if !data.Geometry.IsNull() && !data.Geometry.IsUnknown() {
		geometryJSON := data.Geometry.ValueString()
		if geometryJSON != "" {
			var rawGeometry map[string]interface{}
			if err := json.Unmarshal([]byte(geometryJSON), &rawGeometry); err != nil {
				return flespi_geofence.Geofence{}, diag.NewErrorDiagnostic(
					"Invalid geometry JSON",
					fmt.Sprintf("Failed to parse geometry: %s", err),
				)
			}

			// Re-marshal to use with UnmarshalGeometry
			geometryBytes, _ := json.Marshal(rawGeometry)
			var err error
			geometry, err = flespi_geofence.UnmarshalGeometry(geometryBytes)
			if err != nil {
				return flespi_geofence.Geofence{}, diag.NewErrorDiagnostic(
					"Invalid geometry format",
					fmt.Sprintf("Failed to unmarshal geometry: %s", err),
				)
			}
		}
	}

	return flespi_geofence.Geofence{
		Id:       data.ID.ValueInt64(),
		Name:     data.Name.ValueString(),
		Enabled:  data.Enabled.ValueBool(),
		Priority: data.Priority.ValueInt64(),
		Geometry: geometry,
	}, nil
}

func (g *gwGeofenceResource) convertFlespiGeofenceToResourceModel(data flespi_geofence.Geofence) (*geofenceResourceModel, diag.Diagnostic) {
	var result geofenceResourceModel

	result.ID = types.Int64Value(data.Id)
	result.Name = types.StringValue(data.Name)
	result.Enabled = types.BoolValue(data.Enabled)
	result.Priority = types.Int64Value(data.Priority)

	// Convert geometry to JSON
	if data.Geometry != nil {
		geometryBytes, err := json.Marshal(data.Geometry)
		if err != nil {
			return nil, diag.NewErrorDiagnostic(
				"Failed to marshal geometry",
				fmt.Sprintf("Error converting geometry to JSON: %s", err),
			)
		}
		result.Geometry = jsontypes.NewNormalizedValue(string(geometryBytes))
	} else {
		result.Geometry = jsontypes.NewNormalizedNull()
	}

	return &result, nil
}
