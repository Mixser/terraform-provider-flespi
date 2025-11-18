package storage

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mixser/flespi-client"
	flespi_cdn "github.com/mixser/flespi-client/resources/storage/cdn"
)

var (
	_ resource.Resource              = &cdnResource{}
	_ resource.ResourceWithConfigure = &cdnResource{}
)

type cdnResource struct {
	client *flespi.Client
}

type cdnResourceModel struct {
	Id      types.Int64  `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Size    types.Int64  `tfsdk:"size"`
	Blocked types.Bool   `tfsdk:"blocked"`
}

func NewCDNResource() resource.Resource {
	return &cdnResource{}
}

func (p *cdnResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	client, ok := request.ProviderData.(*flespi.Client)

	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *flespi.Client, got %T. Please report this issue to the provider developers.", request.ProviderData))
		return
	}

	p.client = client
}

func (p *cdnResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_cdn"
}

func (p *cdnResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
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
				Description: "Name of the CDN",
			},
			"size": schema.Int64Attribute{
				Computed:    true,
				Description: "Size of the CDN storage in bytes",
			},
			"blocked": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the CDN is blocked",
			},
		},
	}
}

func (p *cdnResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data *cdnResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	cdnInstance, err := flespi_cdn.NewCDN(p.client, data.Name.ValueString())

	if err != nil {
		response.Diagnostics.AddError(
			"Failed to create CDN",
			fmt.Sprintf("Error creating CDN: %s", err),
		)
		return
	}

	result := p.convertFlespiCDNToResourceModel(cdnInstance)

	response.Diagnostics.Append(response.State.Set(ctx, &result)...)
}

func (p *cdnResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state cdnResourceModel

	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	cdn, err := flespi_cdn.GetCDN(p.client, state.Id.ValueInt64())

	if err != nil {
		response.Diagnostics.AddError(
			"Error Reading Flespi CDN",
			"Could not read Flespi CDN ID "+state.Id.String()+": "+err.Error(),
		)

		return
	}

	result := p.convertFlespiCDNToResourceModel(cdn)

	diags = response.State.Set(ctx, result)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}
}

func (p *cdnResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan cdnResourceModel

	diags := request.Plan.Get(ctx, &plan)

	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	cdn := p.convertResourceModelToFlespiCDN(plan)

	_, err := flespi_cdn.UpdateCDN(p.client, cdn)

	if err != nil {
		response.Diagnostics.AddError(
			"Error Updating Flespi CDN",
			"Could not update CDN, unexpected error: "+err.Error(),
		)
		return
	}

	updatedCDN, err := flespi_cdn.GetCDN(p.client, plan.Id.ValueInt64())

	if err != nil {
		response.Diagnostics.AddError(
			"Error Reading Flespi CDN",
			"Could not read CDN Id: "+plan.Id.String()+": "+err.Error(),
		)
	}

	result := p.convertFlespiCDNToResourceModel(updatedCDN)

	diags = response.State.Set(ctx, result)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}
}

func (p *cdnResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state cdnResourceModel

	diags := request.State.Get(ctx, &state)

	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	err := flespi_cdn.DeleteCDNById(p.client, state.Id.ValueInt64())

	if err != nil {
		response.Diagnostics.AddError(
			"Error Deleting Flespi CDN",
			"Could not delete CDN, unexpected error: "+err.Error(),
		)
		return
	}
}

func (p *cdnResource) convertFlespiCDNToResourceModel(cdn *flespi_cdn.CDN) *cdnResourceModel {
	return &cdnResourceModel{
		Id:      types.Int64Value(cdn.Id),
		Name:    types.StringValue(cdn.Name),
		Size:    types.Int64Value(cdn.Size),
		Blocked: types.BoolValue(cdn.Blocked),
	}
}

func (p *cdnResource) convertResourceModelToFlespiCDN(data cdnResourceModel) flespi_cdn.CDN {
	return flespi_cdn.CDN{
		Id:      data.Id.ValueInt64(),
		Name:    data.Name.ValueString(),
		Size:    data.Size.ValueInt64(),
		Blocked: data.Blocked.ValueBool(),
	}
}
