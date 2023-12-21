package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mixser/flespi-client"
)

var (
	_ resource.Resource              = &platformLimitResource{}
	_ resource.ResourceWithConfigure = &platformLimitResource{}
)

func NewLimitResource() resource.Resource {
	return &platformLimitResource{}
}

type platformLimitResource struct {
	client *flespi.Client
}

type limitResourceModel struct {
	Id               types.Int64  `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	BlockingDuration types.Int64  `tfsdk:"blocking_duration"`
	ApiCall          types.Int64  `tfsdk:"api_calls"`
	ApiTraffic       types.Int64  `tfsdk:"api_traffic"`

	ChannelsCount      types.Int64 `tfsdk:"channels_count"`
	ChannelMessages    types.Int64 `tfsdk:"channel_messages"`
	ChannelStorage     types.Int64 `tfsdk:"channel_storage"`
	ChannelTraffic     types.Int64 `tfsdk:"channel_traffic"`
	ChannelConnections types.Int64 `tfsdk:"channel_connections"`

	ContainersCount  types.Int64 `tfsdk:"containers_count"`
	ContainerStorage types.Int64 `tfsdk:"container_storage"`

	CdnsCount  types.Int64 `tfsdk:"cdns_count"`
	CdnStorage types.Int64 `tfsdk:"cdn_storage"`
	CdnTraffic types.Int64 `tfsdk:"cdn_traffic"`

	DevicesCount       types.Int64 `tfsdk:"devices_count"`
	DeviceStorage      types.Int64 `tfsdk:"device_storage"`
	DeviceMediaTraffic types.Int64 `tfsdk:"device_media_traffic"`
	DeviceMediaStorage types.Int64 `tfsdk:"device_media_storage"`

	StreamsCount  types.Int64 `tfsdk:"streams_count"`
	StreamStorage types.Int64 `tfsdk:"stream_storage"`
	StreamTraffic types.Int64 `tfsdk:"stream_traffic"`

	ModemsCount types.Int64 `tfsdk:"modems_count"`

	MqttSessions        types.Int64 `tfsdk:"mqtt_sessions"`
	MqttMessages        types.Int64 `tfsdk:"mqtt_messages"`
	MqttSessionStorage  types.Int64 `tfsdk:"mqtt_session_storage"`
	MqttRetainedStorage types.Int64 `tfsdk:"mqtt_retained_storage"`
	MqttSubscriptions   types.Int64 `tfsdk:"mqtt_subscriptions"`

	SmsCount types.Int64 `tfsdk:"sms_count"`

	TokensCount types.Int64 `tfsdk:"tokens_count"`

	SubaccountsCount types.Int64 `tfsdk:"subaccounts_count"`

	LimitsCount types.Int64 `tfsdk:"limits_count"`

	RealmsCount types.Int64 `tfsdk:"realms_count"`

	CalcsCount   types.Int64 `tfsdk:"calcs_count"`
	CalcsStorage types.Int64 `tfsdk:"calcs_storage"`

	PluginsCount           types.Int64 `tfsdk:"plugins_count"`
	PluginTraffic          types.Int64 `tfsdk:"plugin_traffic"`
	PluginBufferedMessages types.Int64 `tfsdk:"plugin_buffered_messages"`

	GroupsCount types.Int64 `tfsdk:"groups_count"`

	WebhooksCount  types.Int64 `tfsdk:"webhooks_count"`
	WebhookStorage types.Int64 `tfsdk:"webhook_storage"`
	WebhookTraffic types.Int64 `tfsdk:"webhook_traffic"`

	GrantsCount types.Int64 `tfsdk:"grants_count"`

	IdentityProvidersCount types.Int64 `tfsdk:"identity_providers_count"`
}

func (p *platformLimitResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_limit"
}

func (p *platformLimitResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*flespi.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *flespi.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	p.client = client
}

func (p *platformLimitResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
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
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"blocking_duration": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(60),
			},
			"api_calls": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(20),
			},
			"api_traffic": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},
			"channels_count": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},
			"channel_messages": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},
			"channel_storage": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},
			"channel_traffic": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},
			"channel_connections": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},
			"containers_count": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},
			"container_storage": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},
			"cdns_count": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},
			"cdn_storage": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},
			"cdn_traffic": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},

			"devices_count": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},
			"device_storage": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},
			"device_media_traffic": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},
			"device_media_storage": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},

			"streams_count": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},
			"stream_storage": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},
			"stream_traffic": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},
			"modems_count": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},

			"mqtt_sessions": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},
			"mqtt_messages": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},
			"mqtt_session_storage": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},
			"mqtt_retained_storage": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},
			"mqtt_subscriptions": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},

			"sms_count": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},

			"tokens_count": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},

			"subaccounts_count": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},

			"limits_count": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},

			"realms_count": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},

			"calcs_count": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},
			"calcs_storage": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},

			"plugins_count": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},
			"plugin_traffic": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},
			"plugin_buffered_messages": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},

			"groups_count": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},

			"webhooks_count": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},
			"webhook_storage": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},
			"webhook_traffic": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},

			"grants_count": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},

			"identity_providers_count": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},
		},
	}
}

func (p *platformLimitResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data *limitResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	newLimitInstance := p.convertResourceModelToFlespiLimit(*data)

	limitInstance, err := p.client.NewLimit(newLimitInstance)

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("%s", err))
		return
	}

	data.Id = types.Int64Value(limitInstance.Id)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (p *platformLimitResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state limitResourceModel

	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	limit, err := p.client.GetLimit(state.Id.ValueInt64())

	if err != nil {
		response.Diagnostics.AddError(
			"Error Reading Flespi Limits",
			"Could not read Flespi limit ID "+state.Id.String()+": "+err.Error(),
		)

		return
	}

	diags = response.State.Set(ctx, p.convertFlespiLimitToResourceModel(limit))

	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}
}

func (p *platformLimitResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan limitResourceModel

	diags := request.Plan.Get(ctx, &plan)

	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	var limit = p.convertResourceModelToFlespiLimit(plan)

	_, err := p.client.UpdateLimit(plan.Id.ValueInt64(), limit)

	if err != nil {
		response.Diagnostics.AddError(
			"Error Updating Flespi Limit",
			"Could not update limit, unexpected error: "+err.Error(),
		)
		return
	}

	updatedLimit, err := p.client.GetLimit(plan.Id.ValueInt64())

	if err != nil {
		response.Diagnostics.AddError(
			"Error Reading Flespi Limit",
			"Could not read limit Id: "+plan.Id.String()+": "+err.Error(),
		)
	}

	diags = response.State.Set(ctx, p.convertFlespiLimitToResourceModel(updatedLimit))
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

}

func (p *platformLimitResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	//TODO implement me
	var state limitResourceModel

	diags := request.State.Get(ctx, &state)

	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	err := p.client.DeleteLimit(state.Id.ValueInt64())

	if err != nil {
		response.Diagnostics.AddError(
			"Error Deleting Flespi Limit",
			"Could not delete limit, unexpected error: "+err.Error(),
		)
		return
	}
}

func (p *platformLimitResource) convertFlespiLimitToResourceModel(limit *flespi.Limit) *limitResourceModel {
	var state limitResourceModel

	state.Id = types.Int64Value(limit.Id)
	state.Name = types.StringValue(limit.Name)
	state.Description = types.StringValue(limit.Description)
	state.BlockingDuration = types.Int64Value(int64(limit.BlockingDuration))

	state.ApiCall = types.Int64Value(limit.ApiCall)
	state.ApiTraffic = types.Int64Value(limit.ApiTraffic)

	state.ChannelsCount = types.Int64Value(limit.ChannelsCount)
	state.ChannelMessages = types.Int64Value(limit.ChannelMessages)
	state.ChannelStorage = types.Int64Value(limit.ChannelStorage)
	state.ChannelTraffic = types.Int64Value(limit.ChannelTraffic)
	state.ChannelConnections = types.Int64Value(limit.ChannelConnections)

	state.ContainersCount = types.Int64Value(limit.ContainersCount)
	state.ContainerStorage = types.Int64Value(limit.ContainerStorage)

	state.CdnsCount = types.Int64Value(limit.CdnsCount)
	state.CdnStorage = types.Int64Value(limit.CdnStorage)
	state.CdnTraffic = types.Int64Value(limit.CdnTraffic)

	state.DevicesCount = types.Int64Value(limit.DevicesCount)
	state.DeviceStorage = types.Int64Value(limit.DeviceStorage)
	state.DeviceMediaTraffic = types.Int64Value(limit.DeviceMediaTraffic)
	state.DeviceMediaStorage = types.Int64Value(limit.DeviceMediaStorage)

	state.StreamsCount = types.Int64Value(limit.StreamsCount)
	state.StreamStorage = types.Int64Value(limit.StreamStorage)
	state.StreamTraffic = types.Int64Value(limit.StreamTraffic)

	state.ModemsCount = types.Int64Value(limit.ModemsCount)

	state.MqttSessions = types.Int64Value(limit.MqttSessions)
	state.MqttMessages = types.Int64Value(limit.MqttMessages)
	state.MqttSessionStorage = types.Int64Value(limit.MqttSessionStorage)
	state.MqttRetainedStorage = types.Int64Value(limit.MqttRetainedStorage)
	state.MqttSubscriptions = types.Int64Value(limit.MqttSubscriptions)

	state.SmsCount = types.Int64Value(limit.SmsCount)

	state.TokensCount = types.Int64Value(limit.TokensCount)

	state.SubaccountsCount = types.Int64Value(limit.SubaccountsCount)

	state.LimitsCount = types.Int64Value(limit.LimitsCount)

	state.RealmsCount = types.Int64Value(limit.RealmsCount)

	state.CalcsCount = types.Int64Value(limit.CalcsCount)
	state.CalcsStorage = types.Int64Value(limit.CalcsStorage)

	state.PluginsCount = types.Int64Value(limit.PluginsCount)
	state.PluginTraffic = types.Int64Value(limit.PluginTraffic)
	state.PluginBufferedMessages = types.Int64Value(limit.PluginBufferedMessages)

	state.GroupsCount = types.Int64Value(limit.GroupsCount)

	state.WebhooksCount = types.Int64Value(limit.WebhooksCount)
	state.WebhookStorage = types.Int64Value(limit.WebhookStorage)
	state.WebhookTraffic = types.Int64Value(limit.WebhookTraffic)

	state.GrantsCount = types.Int64Value(limit.GrantsCount)
	state.IdentityProvidersCount = types.Int64Value(limit.IdentityProvidersCount)

	return &state
}

func (p *platformLimitResource) convertResourceModelToFlespiLimit(data limitResourceModel) flespi.Limit {
	return flespi.Limit{
		Id:                     data.Id.ValueInt64(),
		Name:                   data.Name.ValueString(),
		Description:            data.Description.ValueString(),
		BlockingDuration:       int(data.BlockingDuration.ValueInt64()),
		ApiCall:                data.ApiCall.ValueInt64(),
		ApiTraffic:             data.ApiTraffic.ValueInt64(),
		ChannelsCount:          data.ChannelsCount.ValueInt64(),
		ChannelMessages:        data.ChannelMessages.ValueInt64(),
		ChannelStorage:         data.ChannelStorage.ValueInt64(),
		ChannelTraffic:         data.ChannelTraffic.ValueInt64(),
		ChannelConnections:     data.ChannelConnections.ValueInt64(),
		ContainersCount:        data.ContainersCount.ValueInt64(),
		ContainerStorage:       data.ContainerStorage.ValueInt64(),
		CdnsCount:              data.CdnsCount.ValueInt64(),
		CdnStorage:             data.CdnStorage.ValueInt64(),
		CdnTraffic:             data.CdnTraffic.ValueInt64(),
		DevicesCount:           data.DevicesCount.ValueInt64(),
		DeviceStorage:          data.DeviceStorage.ValueInt64(),
		DeviceMediaTraffic:     data.DeviceMediaTraffic.ValueInt64(),
		DeviceMediaStorage:     data.DeviceMediaStorage.ValueInt64(),
		StreamsCount:           data.StreamsCount.ValueInt64(),
		StreamStorage:          data.StreamStorage.ValueInt64(),
		StreamTraffic:          data.StreamTraffic.ValueInt64(),
		ModemsCount:            data.ModemsCount.ValueInt64(),
		MqttSessions:           data.MqttSessions.ValueInt64(),
		MqttMessages:           data.MqttMessages.ValueInt64(),
		MqttSessionStorage:     data.MqttSessionStorage.ValueInt64(),
		MqttRetainedStorage:    data.MqttRetainedStorage.ValueInt64(),
		MqttSubscriptions:      data.MqttSubscriptions.ValueInt64(),
		SmsCount:               data.SmsCount.ValueInt64(),
		TokensCount:            data.TokensCount.ValueInt64(),
		SubaccountsCount:       data.SubaccountsCount.ValueInt64(),
		LimitsCount:            data.LimitsCount.ValueInt64(),
		RealmsCount:            data.RealmsCount.ValueInt64(),
		CalcsCount:             data.CalcsCount.ValueInt64(),
		CalcsStorage:           data.CalcsStorage.ValueInt64(),
		PluginsCount:           data.PluginsCount.ValueInt64(),
		PluginTraffic:          data.PluginTraffic.ValueInt64(),
		PluginBufferedMessages: data.PluginBufferedMessages.ValueInt64(),
		GroupsCount:            data.GroupsCount.ValueInt64(),
		WebhooksCount:          data.WebhooksCount.ValueInt64(),
		WebhookStorage:         data.WebhookStorage.ValueInt64(),
		WebhookTraffic:         data.WebhookTraffic.ValueInt64(),
		GrantsCount:            data.GrantsCount.ValueInt64(),
		IdentityProvidersCount: data.IdentityProvidersCount.ValueInt64(),
		Metadata:               map[string]string{},
	}
}
