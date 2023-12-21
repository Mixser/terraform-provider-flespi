package provider

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
)

var (
	_ resource.Resource              = &platformSubaccountResource{}
	_ resource.ResourceWithConfigure = &platformSubaccountResource{}
)

func NewSubaccountResource() resource.Resource {
	return &platformSubaccountResource{}
}

type platformSubaccountResource struct {
	client *flespi.Client
}

type subaccountResourceModel struct {
	Id      types.Int64  `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	LimitId types.Int64  `tfsdk:"limit_id"`
}

func (p *platformSubaccountResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (p *platformSubaccountResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_subaccount"
}

func (p *platformSubaccountResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
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
			"limit_id": schema.Int64Attribute{
				Required: true,
			},
		},
	}
}

func (p *platformSubaccountResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data *subaccountResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	newSubaccountInstance := flespi.Subaccount{
		Name:     data.Name.ValueString(),
		LimitId:  data.LimitId.ValueInt64(),
		Metadata: map[string]string{},
	}

	subaccountInstatnce, err := p.client.NewSubaccount(newSubaccountInstance)

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("%s", err))
		return
	}

	data.Id = types.Int64Value(subaccountInstatnce.Id)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (p *platformSubaccountResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state subaccountResourceModel

	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	subaccount, err := p.client.GetSubaccount(state.Id.ValueInt64())

	if err != nil {
		response.Diagnostics.AddError(
			"Error Reading Flespi Subaccounts",
			"Could not read Flespi subaccount ID "+state.Id.String()+": "+err.Error(),
		)

		return
	}

	diags = response.State.Set(ctx, p.convertFlespiSubaccountToResourceModel(subaccount))

	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

}

func (p *platformSubaccountResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan subaccountResourceModel

	diags := request.Plan.Get(ctx, &plan)

	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	subaccount := p.convertResourceModelToFlespiSubaccount(plan)

	_, err := p.client.UpdateSubaccount(plan.Id.ValueInt64(), subaccount)

	if err != nil {
		response.Diagnostics.AddError(
			"Error Updating Flespi Subaccount",
			"Could not update subaccount, unexpected error: "+err.Error(),
		)
		return
	}

	updatedSubaccount, err := p.client.GetSubaccount(plan.Id.ValueInt64())

	if err != nil {
		response.Diagnostics.AddError(
			"Error Reading Flespi Subaccount",
			"Could not read subaccount Id: "+plan.Id.String()+": "+err.Error(),
		)
	}

	diags = response.State.Set(ctx, p.convertFlespiSubaccountToResourceModel(updatedSubaccount))
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}
}

func (p *platformSubaccountResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state subaccountResourceModel

	diags := request.State.Get(ctx, &state)

	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	err := p.client.DeleteSubaccount(state.Id.ValueInt64())

	if err != nil {
		response.Diagnostics.AddError(
			"Error Deleting Flespi Subaccount",
			"Could not delete subaccount, unexpected error: "+err.Error(),
		)
		return
	}
}

func (p *platformSubaccountResource) convertFlespiSubaccountToResourceModel(subaccount *flespi.Subaccount) *subaccountResourceModel {
	return &subaccountResourceModel{
		Id:      types.Int64Value(subaccount.Id),
		Name:    types.StringValue(subaccount.Name),
		LimitId: types.Int64Value(subaccount.LimitId),
	}
}

func (p *platformSubaccountResource) convertResourceModelToFlespiSubaccount(data subaccountResourceModel) flespi.Subaccount {
	return flespi.Subaccount{
		Id:       data.Id.ValueInt64(),
		Name:     data.Name.ValueString(),
		LimitId:  data.LimitId.ValueInt64(),
		Metadata: map[string]string{},
	}
}
