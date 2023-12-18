package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mixser/flespi-client"
)

var (
	_ resource.Resource              = &platformCdnResource{}
	_ resource.ResourceWithConfigure = &platformCdnResource{}
)

func NewCdnResource() resource.Resource {
	return &platformCdnResource{}
}

type platformCdnResource struct {
	client *flespi.Client
}

func (d *platformCdnResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	d.client = client
}

func (d *platformCdnResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cdn"
}

func (d *platformCdnResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

type cdnsDataSourceModel struct {
	CDNS []cdnModel `tfsdk:"cdns"`
}

type cdnModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *platformCdnResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
}

func (r *platformCdnResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	//	var state cdnModel
	//
	//	diags := req.State.Get(ctx, &state)
	//
	//	resp.Diagnostics.Append(diags...)
	//	if resp.Diagnostics.HasError() {
	//		return
	//	}
	//
	//	order, err := r.client.GetCDNS(state.ID.ValueString())
	//
	//	cdns, err := d.client.GetCDNS()
	//
	//	if err != nil {
	//		resp.Diagnostics.AddError(
	//			"Unable to Read Flespi CDNS",
	//			err.Error(),
	//		)
	//	}
	//
	//	for _, cdn := range cdns {
	//		cdnState := cdnModel{
	//			ID: types.Int64Value(int64(cdn.Id)),
	//			Name: types.StringValue(cdn.Name),
	//		}
	//
	//		state.CDNS = append(state.CDNS, cdnState)
	//	}
	//
	//	diags := resp.State.Set(ctx, &state)
	//
	//	resp.Diagnostics.Append(diags...)
	//	if resp.Diagnostics.HasError() {
	//		return
	//	}
	return
}

func (r *platformCdnResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *platformCdnResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

}
