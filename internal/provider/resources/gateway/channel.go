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
	flespi "github.com/mixser/flespi-client"
	flespi_channel "github.com/mixser/flespi-client/resources/gateway/channel"
)

type gwChannelResource struct {
	client *flespi_channel.ChannelClient
}

type channelResourceModel struct {
	Id   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`

	Enabled types.Bool `tfsdk:"enabled"`

	ProtocolId   types.Int64  `tfsdk:"protocol_id"`
	ProtocolName types.String `tfsdk:"protocol_name"`

	MessagesTTL types.Int64 `tfsdk:"messages_ttl"`

	Configuration jsontypes.Normalized `tfsdk:"configuration"`
	Metadata      types.Map            `tfsdk:"metadata"`
	AccountId     types.Int64          `tfsdk:"account_id"`
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
				Optional:    true,
				Computed:    true,
				CustomType:  jsontypes.NormalizedType{},
				Description: "Protocol-specific configuration parameters as JSON. The available fields depend on the protocol; retrieve the schema with GET /gw/channel-protocols/{protocol_id}. Use jsonencode() in HCL.",
			},
			"metadata": schema.MapAttribute{
				Optional:    true,
				Computed:    false,
				ElementType: types.StringType,
			},
			"account_id": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Subaccount ID to create the channel under. Passed via x-flespi-cid header, not stored in the channel itself.",
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

	g.client = client.Channels
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

	createOpts := []flespi_channel.CreateChannelOption{
		flespi_channel.WithStatus(instance.Enabled),
		flespi_channel.WithMessagesTTL(instance.MessagesTTL),
		flespi_channel.WithConfiguration(instance.Configuration),
		flespi_channel.WithMetadata(instance.Metadata),
	}
	if instance.AccountId != 0 {
		createOpts = append(createOpts, flespi_channel.WithAccountId(instance.AccountId))
	}

	if instance.ProtocolId != 0 {
		channelInstance, err = g.client.CreateWithProtocolId(instance.Name, instance.ProtocolId, createOpts...)
	} else if instance.ProtocolName != "" {
		channelInstance, err = g.client.CreateWithProtocolName(instance.Name, instance.ProtocolName, createOpts...)
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

	channelInstance, err = g.client.Get(channelInstance.Id)

	if err != nil {
		response.Diagnostics.AddError(
			"Failed to get channel after creation",
			fmt.Sprintf("Error creating channel: %s", err),
		)
		return
	}

	result, diags := g.convertFlespiChannelToResourceModel(ctx, channelInstance)

	response.Diagnostics.Append(diags...)
	response.Diagnostics.Append(response.State.Set(ctx, &result)...)
}

func (g *gwChannelResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state channelResourceModel

	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	channelInstance, err := g.client.Get(state.Id.ValueInt64())

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
	var plan channelResourceModel
	var state channelResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	// Id is Computed and not present in the plan; carry it over from state.
	plan.Id = state.Id

	channelInstance, diags := g.convertResourceModelToFlespiChannel(ctx, plan)

	if diags != nil && diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	_, err := g.client.Update(channelInstance)

	if err != nil {
		response.Diagnostics.AddError(
			"Failed to update channel",
			fmt.Sprintf("Error updating channel: %s", err),
		)
		return
	}

	updatedChannel, err := g.client.Get(state.Id.ValueInt64())
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

	err := g.client.DeleteById(data.Id.ValueInt64())

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

	if !data.Configuration.IsNull() && !data.Configuration.IsUnknown() {
		if err := json.Unmarshal([]byte(data.Configuration.ValueString()), &configuration); err != nil {
			return flespi_channel.Channel{}, diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unable to unmarshal configuration: "+data.Configuration.String(),
					"Unable to unmarshal configuration: "+err.Error(),
				),
			}
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
		AccountId:     data.AccountId.ValueInt64(),
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

	data.Configuration = jsontypes.NewNormalizedValue(string(configuration))

	metadata, diags := types.MapValueFrom(ctx, types.StringType, channel.Metadata)
	if diags.HasError() {
		return nil, diags
	}

	data.Metadata = metadata
	data.AccountId = types.Int64Value(channel.AccountId)

	return &data, nil
}
