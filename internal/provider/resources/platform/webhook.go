package platform

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mixser/flespi-client"
	flespi_webhook "github.com/mixser/flespi-client/resources/platform/webhook"
)

var (
	_ resource.Resource              = &platformWebhookResource{}
	_ resource.ResourceWithConfigure = &platformWebhookResource{}
)

type platformWebhookResource struct {
	client *flespi.Client
}

type webhookResourceModel struct {
	Id             types.Int64          `tfsdk:"id"`
	Name           types.String         `tfsdk:"name"`
	Type           types.String         `tfsdk:"type"`
	Triggers       []triggerModel       `tfsdk:"triggers"`
	Configurations []configurationModel `tfsdk:"configurations"`
}

type header struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type validator struct {
	Expression types.String `tfsdk:"expression"`
	Action     types.String `tfsdk:"action"`
}

type configurationModel struct {
	Type     types.String `tfsdk:"type"`
	Uri      types.String `tfsdk:"uri"`
	Method   types.String `tfsdk:"method"`
	Body     types.String `tfsdk:"body"`
	CA       types.String `tfsdk:"ca"`
	CID      types.String `tfsdk:"cid"`
	Headers  []header     `tfsdk:"headers"`
	Validate *validator   `tfsdk:"validate"`
}

type filterModel struct {
	CID     types.Int64  `tfsdk:"cid"`
	Payload types.String `tfsdk:"payload"`
}

type triggerModel struct {
	Topic  types.String `tfsdk:"topic"`
	Filter *filterModel `tfsdk:"filter"`
}

func NewWebhookResource() resource.Resource {
	return &platformWebhookResource{}
}

func (p *platformWebhookResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	client, ok := request.ProviderData.(*flespi.Client)

	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *flespi.Client, got %T.Please report this issue to the provider developers.", request.ProviderData))
		return
	}

	p.client = client
}

func (p platformWebhookResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_webhook"
}

func (p platformWebhookResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	//TODO implement me
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"triggers": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"topic": schema.StringAttribute{
							Required: true,
						},
						"filter": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"cid": schema.Int64Attribute{
									Computed: false,
									Optional: true,
								},
								"payload": schema.StringAttribute{
									Required: true,
								},
							},
							Optional: true,
						},
					},
				},
				Required: true,
			},
			"configurations": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required: true,
						},
						"uri": schema.StringAttribute{
							Required: true,
						},
						"method": schema.StringAttribute{
							Required: true,
						},
						"body": schema.StringAttribute{
							Optional: true,
						},
						"ca": schema.StringAttribute{
							Optional: true,
						},
						"cid": schema.StringAttribute{
							Optional: true,
						},
						"headers": schema.ListNestedAttribute{
							Required: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name":  schema.StringAttribute{Required: true},
									"value": schema.StringAttribute{Required: true},
								},
							},
						},
						"validate": schema.SingleNestedAttribute{
							Optional: true,
							Attributes: map[string]schema.Attribute{
								"action": schema.StringAttribute{
									Required: true,
								},
								"expression": schema.StringAttribute{
									Required: true,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (p platformWebhookResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data *webhookResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	newWebhookInstance := convertWebhookResourceModelToFlespiWebhook(*data)

	var webhookInstance flespi_webhook.Webhook
	var err error

	switch newWebhookInstance.(type) {
	case *flespi_webhook.SingleWebhook:
		webhook := *newWebhookInstance.(*flespi_webhook.SingleWebhook)
		webhookInstance, err = flespi_webhook.NewSignleWebhook(
			p.client,
			webhook.Name,
			flespi_webhook.SWWithTriggers(webhook.Triggers),
			flespi_webhook.SWWithConfiguration(webhook.Configuration),
		)
	case *flespi_webhook.ChainedWebhook:
		webhook := *newWebhookInstance.(*flespi_webhook.ChainedWebhook)
		webhookInstance, err = flespi_webhook.NewChainedWebhook(
			p.client,
			webhook.Name,
			flespi_webhook.CWWithTriggers(webhook.Triggers),
			flespi_webhook.CWWithConfigurations(webhook.Configuration),
		)
	}

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("%s", err))
		return
	}

	data.Id = convertFlespiWebhookToResourceModel(webhookInstance).Id

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (p platformWebhookResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state webhookResourceModel

	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	webhook, err := flespi_webhook.GetWebhook(p.client, state.Id.ValueInt64())

	if err != nil {
		response.Diagnostics.AddError(
			"Error Reading Flespi Webhooks",
			"Could not read Flespi Webhook ID "+state.Id.String()+": "+err.Error(),
		)

		return
	}

	diags = response.State.Set(ctx, convertFlespiWebhookToResourceModel(webhook))

	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

}

func (p platformWebhookResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan webhookResourceModel

	diags := request.Plan.Get(ctx, &plan)

	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	webhookId := plan.Id.ValueInt64()
	plan.Id = types.Int64Value(0)

	webhook := convertWebhookResourceModelToFlespiWebhook(plan)

	_, err := flespi_webhook.UpdateWebhook(p.client, webhook)

	if err != nil {
		response.Diagnostics.AddError(
			"Error Updating Flespi Webhook",
			"Could not update webhook, unexpected error: "+err.Error(),
		)
		return
	}

	updatedWebhook, err := flespi_webhook.GetWebhook(p.client, webhookId)

	if err != nil {
		response.Diagnostics.AddError(
			"Error Reading Flespi Subaccount",
			fmt.Sprintf("Could not read subaccount Id: %d: %s", webhookId, err.Error()),
		)
	}

	diags = response.State.Set(ctx, convertFlespiWebhookToResourceModel(updatedWebhook))
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}
}

func (p platformWebhookResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state webhookResourceModel

	diags := request.State.Get(ctx, &state)

	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	err := flespi_webhook.DeleteWebhookById(p.client, state.Id.ValueInt64())

	if err != nil {
		response.Diagnostics.AddError(
			"Error Deleting Flespi Webhook",
			"Could not delete webhook, unexpected error: "+err.Error(),
		)
		return
	}
}

func convertFlespiWebhookToResourceModel(webhook flespi_webhook.Webhook) *webhookResourceModel {
	switch v := webhook.(type) {
	case *flespi_webhook.SingleWebhook:
		return convertFlespiSingleWebhookToResourceModel(v)
	case *flespi_webhook.ChainedWebhook:
		return convertFlespiChainedWebhookToResourceModel(v)
	default:
		panic(fmt.Sprintf("Unknown type: %T", webhook))
	}
}

func convertFlespiSingleWebhookToResourceModel(webhook *flespi_webhook.SingleWebhook) *webhookResourceModel {
	var result = webhookResourceModel{
		Id:             types.Int64Value(webhook.Id),
		Name:           types.StringValue(webhook.Name),
		Type:           types.StringValue("single-webhook"),
		Triggers:       convertFlespiTriggerToResourceModel(webhook.Triggers),
		Configurations: []configurationModel{convertFlespiConfigurationToResourceModel(webhook.Configuration)},
	}

	return &result
}

func convertFlespiChainedWebhookToResourceModel(webhook *flespi_webhook.ChainedWebhook) *webhookResourceModel {
	var result = webhookResourceModel{
		Id:       types.Int64Value(webhook.Id),
		Type:     types.StringValue("chained-webhook"),
		Name:     types.StringValue(webhook.Name),
		Triggers: convertFlespiTriggerToResourceModel(webhook.Triggers),
	}

	for _, cfg := range webhook.Configuration {
		result.Configurations = append(result.Configurations, convertFlespiConfigurationToResourceModel(cfg))
	}

	return &result
}

func convertFlespiTriggerToResourceModel(triggers []flespi_webhook.Trigger) []triggerModel {
	var result = []triggerModel{}

	for _, trigger := range triggers {
		result = append(result, triggerModel{
			Topic:  types.StringValue(trigger.Topic),
			Filter: convertFlespiTriggerFilterToResourceModel(trigger.Filter),
		})
	}
	return result
}

func convertFlespiTriggerFilterToResourceModel(filter *flespi_webhook.TriggerFilter) *filterModel {
	if filter == nil {
		return nil
	}

	return &filterModel{
		CID:     types.Int64Value(filter.CID),
		Payload: types.StringValue(filter.Payload),
	}
}

func convertFlespiConfigurationToResourceModel(configuration flespi_webhook.Configuration) configurationModel {
	switch v := configuration.(type) {
	case flespi_webhook.CustomServerConfiguration:
		return convertFlespiCustomServiceConfigurationToResourceModel(&v)
	case flespi_webhook.FlespiConfiguration:
		return convertFlespiFlespiConfigurationToResourceModel(&v)
	default:
		panic(fmt.Sprintf("Unknown type: %T", configuration))
	}
}

func convertFlespiCustomServiceConfigurationToResourceModel(cfg *flespi_webhook.CustomServerConfiguration) configurationModel {
	var result = configurationModel{
		Type:     types.StringValue(cfg.Type),
		Uri:      types.StringValue(cfg.Uri),
		Method:   types.StringValue(cfg.Method),
		Body:     types.StringValue(cfg.Body),
		CA:       types.StringPointerValue(cfg.CA),
		Headers:  []header{},
		Validate: convertFlespiValidatorToResourceModel(cfg.Validate),
	}

	for _, flespiHeader := range cfg.Headers {
		result.Headers = append(result.Headers, convertFlespiHeaderToResourceModel(flespiHeader))
	}

	return result
}

func convertFlespiFlespiConfigurationToResourceModel(cfg *flespi_webhook.FlespiConfiguration) configurationModel {
	return configurationModel{}
}

func convertFlespiHeaderToResourceModel(flespiHeader flespi_webhook.Header) header {
	return header{
		Name:  types.StringValue(flespiHeader.Name),
		Value: types.StringValue(flespiHeader.Value),
	}
}

func convertFlespiValidatorToResourceModel(v *flespi_webhook.Validator) *validator {
	if v == nil {
		return nil
	}

	return &validator{
		Expression: types.StringValue(v.Expression),
		Action:     types.StringValue(v.Action),
	}
}

func convertWebhookResourceModelToFlespiWebhook(data webhookResourceModel) flespi_webhook.Webhook {
	var result flespi_webhook.Webhook

	switch data.Type.ValueString() {
	case "single-webhook":
		result = &flespi_webhook.SingleWebhook{
			Id:            data.Id.ValueInt64(),
			Name:          data.Name.ValueString(),
			Triggers:      convertTriggersToFlespiTriggers(data.Triggers),
			Configuration: convertConfigurationResourceModelToFlespiConfiguration(data.Configurations[0]),
		}
	case "chained-webhool":
		result = &flespi_webhook.ChainedWebhook{
			Id:            data.Id.ValueInt64(),
			Name:          data.Name.ValueString(),
			Triggers:      convertTriggersToFlespiTriggers(data.Triggers),
			Configuration: convertConfigurationsToFlespiConfigurations(data.Configurations),
		}
	default:
		panic(fmt.Sprintf("Unknown webhook type: %s", data.Type))
	}

	return result
}

func convertConfigurationsToFlespiConfigurations(cfgs []configurationModel) []flespi_webhook.Configuration {
	var result []flespi_webhook.Configuration

	for _, cfg := range cfgs {
		result = append(result, convertConfigurationResourceModelToFlespiConfiguration(cfg))
	}

	return result
}

func convertConfigurationResourceModelToFlespiConfiguration(cfg configurationModel) flespi_webhook.Configuration {
	var result flespi_webhook.Configuration

	switch cfg.Type.ValueString() {
	case "custom-server":
		result = &flespi_webhook.CustomServerConfiguration{
			Type:     cfg.Type.ValueString(),
			Uri:      cfg.Uri.ValueString(),
			Method:   cfg.Method.ValueString(),
			Body:     cfg.Body.ValueString(),
			CA:       cfg.CA.ValueStringPointer(),
			Headers:  convertHeaderstoFlespiHeaders(cfg.Headers),
			Validate: convertValidatorResourceModelToFlespiValidator(cfg.Validate),
		}
	case "flespi-platform":
		result = &flespi_webhook.FlespiConfiguration{
			Type:     cfg.Type.ValueString(),
			Uri:      cfg.Uri.ValueString(),
			Method:   cfg.Method.ValueString(),
			Body:     cfg.Method.ValueString(),
			CID:      cfg.CID.ValueString(),
			Validate: convertValidatorResourceModelToFlespiValidator(cfg.Validate),
		}
	}

	return result
}

func convertHeaderstoFlespiHeaders(hs []header) []flespi_webhook.Header {
	var result []flespi_webhook.Header

	for _, h := range hs {
		result = append(result, convertHeaderResourceModelToFlespiHeader(h))
	}

	return result
}

func convertHeaderResourceModelToFlespiHeader(h header) flespi_webhook.Header {
	return flespi_webhook.Header{
		Name:  h.Name.ValueString(),
		Value: h.Value.ValueString(),
	}
}

func convertValidatorResourceModelToFlespiValidator(v *validator) *flespi_webhook.Validator {
	if v == nil {
		return nil
	}

	return &flespi_webhook.Validator{
		Expression: v.Expression.ValueString(),
		Action:     v.Action.ValueString(),
	}
}

func convertTriggersToFlespiTriggers(ts []triggerModel) []flespi_webhook.Trigger {
	var result []flespi_webhook.Trigger

	for _, t := range ts {

		trigger := flespi_webhook.Trigger{
			Topic: t.Topic.ValueString(),
		}

		if t.Filter != nil {
			trigger.Filter = &flespi_webhook.TriggerFilter{
				CID:     t.Filter.CID.ValueInt64(),
				Payload: t.Filter.Payload.ValueString(),
			}
		}

		result = append(result, trigger)
	}

	return result
}
