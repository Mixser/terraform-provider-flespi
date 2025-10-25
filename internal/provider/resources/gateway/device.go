package gateway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mixser/flespi-client"
	flespi_device "github.com/mixser/flespi-client/resources/gateway/device"
)

var (
	_ resource.Resource              = &gwDeviceResource{}
	_ resource.ResourceWithConfigure = &gwDeviceResource{}
)

type gwDeviceResource struct {
	client *flespi.Client
}

func NewDeviceResource() resource.Resource {
	return &gwDeviceResource{}
}

type deviceResourceModel struct {
	Id   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`

	Enabled types.Bool `tfsdk:"enabled"`

	Configuration types.Map `tfsdk:"configuration"`

	DeviceTypeId types.Int64 `tfsdk:"device_type_id"`

	MessagesTTL    types.Int64 `tfsdk:"messages_ttl"`
	MessagesRotate types.Int64 `tfsdk:"messages_rotate"`

	MediaTTL    types.Int64 `tfsdk:"media_ttl"`
	MediaRotate types.Int64 `tfsdk:"media_rotate"`

	// Metadata types.[string]string `json:"metadata,omitempty"`
}

func (g *gwDeviceResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_device"
}

func (g *gwDeviceResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	client, ok := request.ProviderData.(*flespi.Client)

	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *flespi.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	g.client = client
}

func (g *gwDeviceResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"enabled": schema.BoolAttribute{
				Required: true,
			},
			"device_type_id": schema.Int64Attribute{
				Required: true,
			},
			"messages_ttl": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
			},
			"messages_rotate": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
			},
			"media_ttl": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
			},
			"media_rotate": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
			},
			"configuration": schema.MapAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (g *gwDeviceResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data *deviceResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	instance := g.convertResourceModelToFlespiDevice(ctx, *data)

	deviceInstance, err := flespi_device.NewDevice(
		g.client,
		instance.Name,
		instance.Enabled,
		instance.DeviceTypeId,
		flespi_device.WithMessage(instance.MessagesTTL, instance.MessagesRotate),
		flespi_device.WithMedia(instance.MediaTTL, instance.MediaRotate),
		flespi_device.WithConfiguration(instance.Configuration),
	)

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("%s", err))
		return
	}

	deviceInstance, err = flespi_device.GetDevice(
		g.client,
		deviceInstance.Id,
	)

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("%s", err))
		return
	}

	result, diags := g.convertFlespiDeviceToResourceModel(deviceInstance)

	response.Diagnostics.Append(diags...)

	response.Diagnostics.Append(response.State.Set(ctx, &result)...)
}

func (g *gwDeviceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state deviceResourceModel

	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	device, err := flespi_device.GetDevice(g.client, state.Id.ValueInt64())

	if err != nil {
		response.Diagnostics.AddError(
			"Error Reading Flespi Devices",
			"Could not read Flespi device ID "+state.Id.String()+": "+err.Error(),
		)

		return
	}

	model, diags := g.convertFlespiDeviceToResourceModel(device)

	response.Diagnostics.Append(diags...)
	response.Diagnostics.Append(response.State.Set(ctx, model)...)
}

func (g *gwDeviceResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state deviceResourceModel

	diags := request.Plan.Get(ctx, &state)

	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	var device = g.convertResourceModelToFlespiDevice(ctx, state)

	_, err := flespi_device.UpdateDevice(g.client, device)

	if err != nil {
		response.Diagnostics.AddError(
			"Error Updating Flespi Device",
			"Could not update device, unexpected error: "+err.Error(),
		)
		return
	}

	updatedDevice, err := flespi_device.GetDevice(g.client, state.Id.ValueInt64())

	if err != nil {
		response.Diagnostics.AddError(
			"Error Reading Flespi Limit",
			"Could not read limit Id: "+state.Id.String()+": "+err.Error(),
		)
	}

	model, diags := g.convertFlespiDeviceToResourceModel(updatedDevice)

	response.Diagnostics.Append(diags...)
	response.Diagnostics.Append(response.State.Set(ctx, model)...)
}

func (g *gwDeviceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state deviceResourceModel

	diags := request.State.Get(ctx, &state)

	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	err := flespi_device.DeleteDeviceById(g.client, state.Id.ValueInt64())

	if err != nil {
		response.Diagnostics.AddError(
			"Error Deleting Flespi Device",
			"Could not delete device, unexpected error: "+err.Error(),
		)
		return
	}
}

func (g *gwDeviceResource) convertResourceModelToFlespiDevice(ctx context.Context, data deviceResourceModel) flespi_device.Device {
	configuration := make(map[string]string)

	for key, value := range data.Configuration.Elements() {
		if strValue, ok := value.(types.String); ok {
			configuration[key] = strValue.ValueString()
		} else {
			tflog.Error(ctx, fmt.Sprintf("expected types.String, got %T (skip attribute)", value))
		}
	}

	return flespi_device.Device{
		Id:             data.Id.ValueInt64(),
		Name:           data.Name.ValueString(),
		Enabled:        data.Enabled.ValueBool(),
		DeviceTypeId:   data.DeviceTypeId.ValueInt64(),
		MessagesTTL:    data.MessagesTTL.ValueInt64(),
		MessagesRotate: data.MessagesRotate.ValueInt64(),
		MediaTTL:       data.MediaTTL.ValueInt64(),
		MediaRotate:    data.MediaRotate.ValueInt64(),
		Configuration:  configuration,
		Metadata:       map[string]string{},
	}
}

func (g *gwDeviceResource) convertFlespiDeviceToResourceModel(device *flespi_device.Device) (*deviceResourceModel, diag.Diagnostics) {
	var state deviceResourceModel

	state.Id = types.Int64Value(device.Id)
	state.Name = types.StringValue(device.Name)
	state.Enabled = types.BoolValue(device.Enabled)

	state.DeviceTypeId = types.Int64Value(device.DeviceTypeId)

	state.MessagesTTL = types.Int64Value(device.MessagesTTL)
	state.MessagesRotate = types.Int64Value(device.MessagesRotate)

	state.MediaTTL = types.Int64Value(device.MediaTTL)
	state.MediaRotate = types.Int64Value(device.MediaRotate)

	configuration := make(map[string]attr.Value)

	for key, value := range device.Configuration {
		configuration[key] = types.StringValue(value)
	}

	cfg, diags := types.MapValue(types.StringType, configuration)
	state.Configuration = cfg

	return &state, diags
}
