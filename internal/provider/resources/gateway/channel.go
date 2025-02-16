package gateway

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mixser/flespi-client"
	flespi_channel "github.com/mixser/flespi-client/resources/gateway/channel"
)

type gwChannelResource struct {
	client *flespi.Client
}

type channelResourceModel struct {
	Id   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`

	Enabled types.Bool `tfsdk:"enabled"`

	ProtocolId   types.Int64  `tfsdk:"protocol_id"`
	ProtocolName types.String `tfsdk:"protocol_name"`

	MessagesTTL types.Int64 `tfsdk:"messages_ttl"`

	Configuration types.String `tfsdk:"configuration"`
	Metadata      types.Map    `tfsdk:"metadata"`
}

func NewChannelResource() resource.Resource {
	return &gwChannelResource{}
}

func (g *gwChannelResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_channel"
}

func (g *gwChannelResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"enabled": schema.BoolAttribute{
				Required: true,
			},
			"protocol_id": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"protocol_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"messages_ttl": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"configuration": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"metadata": schema.MapAttribute{
				Optional:    true,
				Computed:    false,
				ElementType: types.StringType,
			},
		},
	}
}

func (g *gwChannelResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (g *gwChannelResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data *channelResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	instance, diags := g.convertResourceModelToFlespiChannel(ctx, *data)

	if diags != nil && diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	var channelInstance *flespi_channel.Channel
	var err error

	if instance.ProtocolId != 0 {
		channelInstance, err = flespi_channel.NewChannelWithProtocolId(
			g.client,
			instance.Name,
			instance.ProtocolId,
			flespi_channel.WithStatus(instance.Enabled),
			flespi_channel.WithMessagesTTL(instance.MessagesTTL),
			flespi_channel.WithConfiguration(instance.Configuration),
			flespi_channel.WithMetadata(instance.Metadata),
		)
	} else if instance.ProtocolName != "" {
		channelInstance, err = flespi_channel.NewChannelWithProtocolName(
			g.client,
			instance.Name,
			instance.ProtocolName,
			flespi_channel.WithStatus(instance.Enabled),
			flespi_channel.WithMessagesTTL(instance.MessagesTTL),
			flespi_channel.WithConfiguration(instance.Configuration),
			flespi_channel.WithMetadata(instance.Metadata),
		)
	} else {
		response.Diagnostics.AddError(
			"Protocol not specified for channel",
			"Please specify either ProtocolId or ProtocolName in the resource configuration.",
		)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			"Failed to create channel",
			fmt.Sprintf("Error creating channel: %s", err),
		)
		return
	}

	channelInstance, err = flespi_channel.GetChannel(
		g.client,
		channelInstance.Id,
	)

	result, diags := g.convertFlespiChannelToResourceModel(ctx, channelInstance)

	response.Diagnostics.Append(diags...)
	response.Diagnostics.Append(response.State.Set(ctx, &result)...)
}

func (g *gwChannelResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data channelResourceModel

	diags := request.State.Get(ctx, &data)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	channelInstance, err := flespi_channel.GetChannel(
		g.client,
		data.Id.ValueInt64(),
	)

	if err != nil {
		response.Diagnostics.AddError(
			"Failed to read channel",
			fmt.Sprintf("Error reading channel: %s", err),
		)
		return
	}

	result, diags := g.convertFlespiChannelToResourceModel(ctx, channelInstance)

	response.Diagnostics.Append(diags...)
	response.Diagnostics.Append(response.State.Set(ctx, &result)...)
}

func (g *gwChannelResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data channelResourceModel

	diags := request.State.Get(ctx, &data)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	channelInstance, diags := g.convertResourceModelToFlespiChannel(ctx, data)

	if diags != nil && diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	_, err := flespi_channel.UpdateChannel(g.client, channelInstance)

	if err != nil {
		response.Diagnostics.AddError(
			"Failed to update channel",
			fmt.Sprintf("Error updating channel: %s", err),
		)
		return
	}

	updatedChannel, err := flespi_channel.GetChannel(g.client, channelInstance.Id)
	if err != nil {
		response.Diagnostics.AddError(
			"Failed to read updated channel",
			fmt.Sprintf("Error reading updated channel: %s", err),
		)
		return
	}

	result, diags := g.convertFlespiChannelToResourceModel(ctx, updatedChannel)
	response.Diagnostics.Append(diags...)
	response.Diagnostics.Append(response.State.Set(ctx, &result)...)
}

func (g *gwChannelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data channelResourceModel

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := flespi_channel.DeleteChannelById(g.client, data.Id.ValueInt64())

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete channel",
			fmt.Sprintf("Error deleting channel: %s", err),
		)
		return
	}
}

func (g *gwChannelResource) convertResourceModelToFlespiChannel(ctx context.Context, data channelResourceModel) (flespi_channel.Channel, diag.Diagnostics) {
	var configuration map[string]interface{}
	metadata := map[string]string{}

	if err := json.Unmarshal([]byte(data.Configuration.ValueString()), &configuration); err != nil {
		return flespi_channel.Channel{}, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Unable to unmarshal configuration: "+data.Configuration.String(),
				"Unable to unmarshal configuration: "+err.Error(),
			),
		}
	}

	if diags := data.Metadata.ElementsAs(ctx, &metadata, false); diags.HasError() {
		return flespi_channel.Channel{}, diags
	}

	return flespi_channel.Channel{
		Id:            data.Id.ValueInt64(),
		Name:          data.Name.ValueString(),
		Enabled:       data.Enabled.ValueBool(),
		ProtocolId:    data.ProtocolId.ValueInt64(),
		ProtocolName:  data.ProtocolName.ValueString(),
		MessagesTTL:   data.MessagesTTL.ValueInt64(),
		Configuration: configuration,
		Metadata:      metadata,
	}, nil
}

func (g *gwChannelResource) convertFlespiChannelToResourceModel(ctx context.Context, channel *flespi_channel.Channel) (*channelResourceModel, diag.Diagnostics) {
	var data channelResourceModel

	data.Id = types.Int64Value(channel.Id)
	data.Name = types.StringValue(channel.Name)
	data.Enabled = types.BoolValue(channel.Enabled)
	data.ProtocolId = types.Int64Value(channel.ProtocolId)
	data.ProtocolName = types.StringValue(channel.ProtocolName)
	data.MessagesTTL = types.Int64Value(channel.MessagesTTL)

	configuration, err := json.Marshal(channel.Configuration)
	if err != nil {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Unable to marshal configuration: "+err.Error(),
				err.Error(),
			),
		}
	}

	data.Configuration = types.StringValue(string(configuration))

	metadata, diags := types.MapValueFrom(ctx, types.StringType, channel.Metadata)
	if diags.HasError() {
		return nil, diags
	}

	data.Metadata = metadata

	return &data, nil
}
