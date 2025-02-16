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
	Blocked bool         `tfsdk:"blocked"`
}

func (p *cdnResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (p *cdnResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_cdns"
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
				Required: true,
			},
			"size": schema.Int64Attribute{
				Required: false,
			},
		},
	}
}

func (p *cdnResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	//TODO implement me
	panic("implement me")
}

func (p *cdnResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	//TODO implement me
	panic("implement me")
}

func (p *cdnResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (p *cdnResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	//TODO implement me
	panic("implement me")
}
