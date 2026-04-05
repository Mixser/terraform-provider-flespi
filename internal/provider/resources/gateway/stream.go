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
	flespi "github.com/mixser/flespi-client"
	flespi_stream "github.com/mixser/flespi-client/resources/gateway/stream"
)

var (
	_ resource.Resource              = &gwStreamResource{}
	_ resource.ResourceWithConfigure = &gwStreamResource{}
)

type gwStreamResource struct {
	client *flespi_stream.StreamClient
}

type streamResourceModel struct {
	Id         types.Int64  `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	ProtocolId types.Int64  `tfsdk:"protocol_id"`

	Enabled  types.Bool  `tfsdk:"enabled"`
	QueueTTL types.Int64 `tfsdk:"queue_ttl"`

	ValidateMessage types.String `tfsdk:"validate_message"`

	Configuration types.Map `tfsdk:"configuration"`
	Metadata      types.Map `tfsdk:"metadata"`
}

func NewStreamResource() resource.Resource {
	return &gwStreamResource{}
}

func (g *gwStreamResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_stream"
}

func (g *gwStreamResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

	g.client = client.Streams
}

func (g *gwStreamResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the stream",
			},
			"protocol_id": schema.Int64Attribute{
				Required:    true,
				Description: "Protocol ID for the stream",
			},
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether the stream is enabled",
			},
			"queue_ttl": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Description: "Queue TTL in seconds",
			},
			"validate_message": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Message validation expression",
			},
			"configuration": schema.MapAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Stream configuration",
			},
			"metadata": schema.MapAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Stream metadata",
			},
		},
	}
}

func (g *gwStreamResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data *streamResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	instance := g.convertResourceModelToFlespiStream(ctx, *data)

	streamInstance, err := g.client.Create(
		instance.Name,
		instance.ProtocolId,
		flespi_stream.WithStatus(instance.Enabled),
		flespi_stream.WithQueueTTL(instance.QueueTTL),
		flespi_stream.WithValidateMessage(instance.ValidateMessage),
		flespi_stream.WithConfiguration(instance.Configuration),
		flespi_stream.WithMetadata(instance.Metadata),
	)

	if err != nil {
		response.Diagnostics.AddError(
			"Failed to create stream",
			fmt.Sprintf("Error creating stream: %s", err),
		)
		return
	}

	streamInstance, err = g.client.Get(streamInstance.Id)

	if err != nil {
		response.Diagnostics.AddError(
			"Failed to get stream after creation",
			fmt.Sprintf("Error reading stream: %s", err),
		)
		return
	}

	result, diags := g.convertFlespiStreamToResourceModel(streamInstance)

	response.Diagnostics.Append(diags...)
	response.Diagnostics.Append(response.State.Set(ctx, &result)...)
}

func (g *gwStreamResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state streamResourceModel

	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	stream, err := g.client.Get(state.Id.ValueInt64())

	if err != nil {
		response.Diagnostics.AddError(
			"Error Reading Flespi Stream",
			"Could not read Flespi stream ID "+state.Id.String()+": "+err.Error(),
		)

		return
	}

	model, diags := g.convertFlespiStreamToResourceModel(stream)

	response.Diagnostics.Append(diags...)
	response.Diagnostics.Append(response.State.Set(ctx, model)...)
}

func (g *gwStreamResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state streamResourceModel

	diags := request.Plan.Get(ctx, &state)

	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	var stream = g.convertResourceModelToFlespiStream(ctx, state)

	_, err := g.client.Update(stream)

	if err != nil {
		response.Diagnostics.AddError(
			"Error Updating Flespi Stream",
			"Could not update stream, unexpected error: "+err.Error(),
		)
		return
	}

	updatedStream, err := g.client.Get(state.Id.ValueInt64())

	if err != nil {
		response.Diagnostics.AddError(
			"Error Reading Flespi Stream",
			"Could not read stream Id: "+state.Id.String()+": "+err.Error(),
		)
	}

	model, diags := g.convertFlespiStreamToResourceModel(updatedStream)

	response.Diagnostics.Append(diags...)
	response.Diagnostics.Append(response.State.Set(ctx, model)...)
}

func (g *gwStreamResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state streamResourceModel

	diags := request.State.Get(ctx, &state)

	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	err := g.client.DeleteById(state.Id.ValueInt64())

	if err != nil {
		response.Diagnostics.AddError(
			"Error Deleting Flespi Stream",
			"Could not delete stream, unexpected error: "+err.Error(),
		)
		return
	}
}

func (g *gwStreamResource) convertResourceModelToFlespiStream(ctx context.Context, data streamResourceModel) flespi_stream.Stream {
	configuration := make(map[string]string)
	metadata := make(map[string]string)

	if !data.Configuration.IsNull() {
		data.Configuration.ElementsAs(ctx, &configuration, false)
	}

	if !data.Metadata.IsNull() {
		data.Metadata.ElementsAs(ctx, &metadata, false)
	}

	return flespi_stream.Stream{
		Id:              data.Id.ValueInt64(),
		Name:            data.Name.ValueString(),
		ProtocolId:      data.ProtocolId.ValueInt64(),
		Enabled:         data.Enabled.ValueBool(),
		QueueTTL:        data.QueueTTL.ValueInt64(),
		ValidateMessage: data.ValidateMessage.ValueString(),
		Configuration:   configuration,
		Metadata:        metadata,
	}
}

func (g *gwStreamResource) convertFlespiStreamToResourceModel(stream *flespi_stream.Stream) (*streamResourceModel, diag.Diagnostics) {
	var state streamResourceModel

	state.Id = types.Int64Value(stream.Id)
	state.Name = types.StringValue(stream.Name)
	state.ProtocolId = types.Int64Value(stream.ProtocolId)
	state.Enabled = types.BoolValue(stream.Enabled)
	state.QueueTTL = types.Int64Value(stream.QueueTTL)
	state.ValidateMessage = types.StringValue(stream.ValidateMessage)

	configuration := make(map[string]attr.Value)
	for key, value := range stream.Configuration {
		configuration[key] = types.StringValue(value)
	}

	cfg, diags := types.MapValue(types.StringType, configuration)
	state.Configuration = cfg

	metadata := make(map[string]attr.Value)
	for key, value := range stream.Metadata {
		metadata[key] = types.StringValue(value)
	}

	meta, metaDiags := types.MapValue(types.StringType, metadata)
	diags.Append(metaDiags...)
	state.Metadata = meta

	return &state, diags
}
