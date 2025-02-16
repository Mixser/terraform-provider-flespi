package provider

import (
	"context"
	"terraform-provider-flespi/internal/provider/resources/gateway"
	"terraform-provider-flespi/internal/provider/resources/platform"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mixser/flespi-client"
)

var _ provider.Provider = &flespiProvider{}

// flespiProvider defines the provider implementation.
type flespiProvider struct {
	version string
}

// FlespiProviderModel describes the provider data model.
type FlespiProviderModel struct {
	Token types.String `tfsdk:"token"`
}

func (p *flespiProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "flespi"
	resp.Version = p.version
}

func (p *flespiProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				MarkdownDescription: "Flespi token",
				Optional:            false,
				Sensitive:           true,
				Required:            true,
			},
		},
	}
}

func (p *flespiProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config FlespiProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown Flespi Token",
			"There is error with Token",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	token := config.Token.ValueString()

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing Flespi Token",
			"We cannot create API client without token",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := flespi.NewClient("https://flespi.io", token)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Flespi API Client",
			"An unexpected error occurred when creating the Flespi API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Flespi Client Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *flespiProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		platform.NewLimitResource,
		platform.NewSubaccountResource,
		platform.NewWebhookResource,
		gateway.NewDeviceResource,
		gateway.NewChannelResource,
	}
}

func (p *flespiProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &flespiProvider{
			version: version,
		}
	}
}
