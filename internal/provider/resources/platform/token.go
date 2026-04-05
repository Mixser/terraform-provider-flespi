package platform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	flespi "github.com/mixser/flespi-client"
	flespi_token "github.com/mixser/flespi-client/resources/gateway/token"
)

var (
	_ resource.Resource              = &platformTokenResource{}
	_ resource.ResourceWithConfigure = &platformTokenResource{}
)

type platformTokenResource struct {
	client *flespi_token.TokenClient
}

type tokenResourceModel struct {
	Id        types.Int64          `tfsdk:"id"`
	Key       types.String         `tfsdk:"key"`
	Info      types.String         `tfsdk:"info"`
	Enabled   types.Bool           `tfsdk:"enabled"`
	Expire    types.Int64          `tfsdk:"expire"`
	TTL       types.Int64          `tfsdk:"ttl"`
	AccountId types.Int64          `tfsdk:"account_id"`
	Metadata  types.Map            `tfsdk:"metadata"`
	Access    jsontypes.Normalized `tfsdk:"access"`
}

func NewTokenResource() resource.Resource {
	return &platformTokenResource{}
}

func (p *platformTokenResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_token"
}

func (p *platformTokenResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

	p.client = client.Tokens
}

func (p *platformTokenResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "Token key (only available after creation)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"info": schema.StringAttribute{
				Required:    true,
				Description: "Token description/info",
			},
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether the token is enabled",
			},
			"expire": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Token expiration timestamp",
			},
			"ttl": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Token TTL in seconds",
			},
			"account_id": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Account ID",
			},
			"metadata": schema.MapAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Token metadata",
			},
			"access": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				CustomType:  jsontypes.NormalizedType{},
				Description: "Token access permissions as JSON. Use jsonencode() in HCL. Example: jsonencode({type=1}) for master, jsonencode({type=0}) for standard, jsonencode({type=2, acl=[{uri=\"gw/devices\", methods=[\"GET\"], ids=\"all\"}]}) for ACL.",
			},
		},
	}
}

func (p *platformTokenResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data *tokenResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	var options []flespi_token.CreateTokenOption

	options = append(options, flespi_token.WithStatus(data.Enabled.ValueBool()))

	if !data.Expire.IsNull() && !data.Expire.IsUnknown() {
		options = append(options, flespi_token.WithExpire(data.Expire.ValueInt64()))
	}

	if !data.TTL.IsNull() && !data.TTL.IsUnknown() {
		options = append(options, flespi_token.WithTTL(data.TTL.ValueInt64()))
	}

	if !data.AccountId.IsNull() && !data.AccountId.IsUnknown() {
		options = append(options, flespi_token.WithAccountId(data.AccountId.ValueInt64()))
	}

	if !data.Access.IsNull() && !data.Access.IsUnknown() {
		var access flespi_token.TokenAccess
		if err := json.Unmarshal([]byte(data.Access.ValueString()), &access); err != nil {
			response.Diagnostics.AddError("Invalid access JSON", err.Error())
			return
		}
		options = append(options, flespi_token.WithAccess(access))
	}

	tokenInstance, err := p.client.Create(data.Info.ValueString(), options...)

	if err != nil {
		response.Diagnostics.AddError(
			"Failed to create token",
			fmt.Sprintf("Error creating token: %s", err),
		)
		return
	}

	result, diags := p.convertFlespiTokenToResourceModel(tokenInstance)

	response.Diagnostics.Append(diags...)
	response.Diagnostics.Append(response.State.Set(ctx, &result)...)
}

func (p *platformTokenResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state tokenResourceModel

	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	token, err := p.client.Get(state.Id.ValueInt64())

	if err != nil {
		response.Diagnostics.AddError(
			"Error Reading Flespi Token",
			"Could not read Flespi token ID "+state.Id.String()+": "+err.Error(),
		)

		return
	}

	// Preserve the key from state since it's not returned by the API after creation
	token.Key = state.Key.ValueString()

	result, diags := p.convertFlespiTokenToResourceModel(token)

	response.Diagnostics.Append(diags...)
	response.Diagnostics.Append(response.State.Set(ctx, result)...)
}

func (p *platformTokenResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan tokenResourceModel
	var state tokenResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	token, diags := p.convertResourceModelToFlespiToken(plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := p.client.Update(token)

	if err != nil {
		response.Diagnostics.AddError(
			"Error Updating Flespi Token",
			"Could not update token, unexpected error: "+err.Error(),
		)
		return
	}

	updatedToken, err := p.client.Get(plan.Id.ValueInt64())

	if err != nil {
		response.Diagnostics.AddError(
			"Error Reading Flespi Token",
			"Could not read token Id: "+plan.Id.String()+": "+err.Error(),
		)
	}

	// Preserve the key from state
	updatedToken.Key = state.Key.ValueString()

	result, diags := p.convertFlespiTokenToResourceModel(updatedToken)

	response.Diagnostics.Append(diags...)
	response.Diagnostics.Append(response.State.Set(ctx, result)...)
}

func (p *platformTokenResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state tokenResourceModel

	diags := request.State.Get(ctx, &state)

	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	err := p.client.DeleteById(state.Id.ValueInt64())

	if err != nil {
		response.Diagnostics.AddError(
			"Error Deleting Flespi Token",
			"Could not delete token, unexpected error: "+err.Error(),
		)
		return
	}
}

func (p *platformTokenResource) convertFlespiTokenToResourceModel(token *flespi_token.Token) (*tokenResourceModel, diag.Diagnostics) {
	var result tokenResourceModel
	var diags diag.Diagnostics

	result.Id = types.Int64Value(token.Id)
	result.Key = types.StringValue(token.Key)
	result.Info = types.StringValue(token.Info)
	result.Enabled = types.BoolValue(token.Enabled)
	result.Expire = types.Int64Value(token.Expire)
	result.TTL = types.Int64Value(token.TTL)
	result.AccountId = types.Int64Value(token.AccountId)

	metadata := make(map[string]attr.Value)
	for key, value := range token.Metadata {
		metadata[key] = types.StringValue(value)
	}
	meta, metaDiags := types.MapValue(types.StringType, metadata)
	diags.Append(metaDiags...)
	result.Metadata = meta

	if token.Access != nil {
		b, err := json.Marshal(token.Access)
		if err != nil {
			diags.AddError("Failed to serialize token access", err.Error())
		} else {
			result.Access = jsontypes.NewNormalizedValue(string(b))
		}
	} else {
		result.Access = jsontypes.NewNormalizedNull()
	}

	return &result, diags
}

func (p *platformTokenResource) convertResourceModelToFlespiToken(data tokenResourceModel) (flespi_token.Token, diag.Diagnostics) {
	var diags diag.Diagnostics
	metadata := make(map[string]string)

	if !data.Metadata.IsNull() {
		for key, value := range data.Metadata.Elements() {
			if strValue, ok := value.(types.String); ok {
				metadata[key] = strValue.ValueString()
			}
		}
	}

	token := flespi_token.Token{
		Id:        data.Id.ValueInt64(),
		Key:       data.Key.ValueString(),
		Info:      data.Info.ValueString(),
		Enabled:   data.Enabled.ValueBool(),
		Expire:    data.Expire.ValueInt64(),
		TTL:       data.TTL.ValueInt64(),
		AccountId: data.AccountId.ValueInt64(),
		Metadata:  metadata,
	}

	if !data.Access.IsNull() && !data.Access.IsUnknown() {
		var access flespi_token.TokenAccess
		if err := json.Unmarshal([]byte(data.Access.ValueString()), &access); err != nil {
			diags.AddError("Invalid access JSON", err.Error())
		} else {
			token.Access = &access
		}
	}

	return token, diags
}
