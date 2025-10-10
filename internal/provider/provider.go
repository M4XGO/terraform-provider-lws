package provider

import (
	"context"
	"os"

	"github.com/M4XGO/terraform-provider-lws/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Nom du provider
const ProviderTypeName = "lws"

// Ensure LWSProvider satisfies various provider interfaces.
var _ provider.Provider = &LWSProvider{}

// LWSProvider defines the provider implementation.
type LWSProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// LWSProviderModel describes the provider data model.
type LWSProviderModel struct {
	Login    types.String `tfsdk:"login"`
	ApiKey   types.String `tfsdk:"api_key"`
	BaseUrl  types.String `tfsdk:"base_url"`
	TestMode types.Bool   `tfsdk:"test_mode"`
	Timeout  types.Int64  `tfsdk:"timeout"`
	Retries  types.Int64  `tfsdk:"retries"`
	Delay    types.Int64  `tfsdk:"delay"`
	Backoff  types.Int64  `tfsdk:"backoff"`
}

func (p *LWSProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = ProviderTypeName
	resp.Version = p.version
}

func (p *LWSProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"login": schema.StringAttribute{
				MarkdownDescription: "LWS login ID. Can also be set with the LWS_LOGIN environment variable.",
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "LWS API key. Can also be set with the LWS_API_KEY environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "LWS API base URL. Defaults to https://api.lws.net/v1. Can also be set with the LWS_BASE_URL environment variable.",
				Optional:            true,
			},
			"test_mode": schema.BoolAttribute{
				MarkdownDescription: "Enable test mode for LWS API. Defaults to false. Can also be set with the LWS_TEST_MODE environment variable.",
				Optional:            true,
			},
			"timeout": schema.Int64Attribute{
				MarkdownDescription: "Timeout for API requests in seconds. Defaults to 30 seconds.",
				Optional:            true,
			},
			"retries": schema.Int64Attribute{
				MarkdownDescription: "Number of retries for API requests. Defaults to 3.",
				Optional:            true,
			},
			"delay": schema.Int64Attribute{
				MarkdownDescription: "Delay between retries for API requests in seconds. Defaults to 15 seconds.",
				Optional:            true,
			},
			"backoff": schema.Int64Attribute{
				MarkdownDescription: "Backoff multiplier for delay between retries. Defaults to 2.",
				Optional:            true,
			},
		},
	}
}

func (p *LWSProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data LWSProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// Example client configuration for data sources and resources
	if data.Login.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("login"),
			"Unknown LWS API Login",
			"The provider cannot create the LWS API client as there is an unknown configuration value for the LWS API login. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the LWS_LOGIN environment variable.",
		)
	}

	if data.ApiKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown LWS API Key",
			"The provider cannot create the LWS API client as there is an unknown configuration value for the LWS API key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the LWS_API_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	login := os.Getenv("LWS_LOGIN")
	apiKey := os.Getenv("LWS_API_KEY")
	baseUrl := os.Getenv("LWS_BASE_URL")
	testMode := os.Getenv("LWS_TEST_MODE") == "true"
	timeout := 30
	retries := 3
	delay := 15
	backoff := 2

	if !data.Login.IsNull() {
		login = data.Login.ValueString()
	}

	if !data.ApiKey.IsNull() {
		apiKey = data.ApiKey.ValueString()
	}

	if !data.BaseUrl.IsNull() {
		baseUrl = data.BaseUrl.ValueString()
	}

	if !data.TestMode.IsNull() {
		testMode = data.TestMode.ValueBool()
	}

	if !data.Timeout.IsNull() {
		timeout = int(data.Timeout.ValueInt64())
	}

	if !data.Retries.IsNull() {
		retries = int(data.Retries.ValueInt64())
	}

	if !data.Delay.IsNull() {
		delay = int(data.Delay.ValueInt64())
	}

	if !data.Backoff.IsNull() {
		backoff = int(data.Backoff.ValueInt64())
	}

	// Default base URL
	if baseUrl == "" {
		baseUrl = "https://api.lws.net/v1"
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if login == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("login"),
			"Missing LWS API Login",
			"The provider requires a LWS login ID. Set the login value in the configuration or use the LWS_LOGIN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing LWS API Key",
			"The provider requires a LWS API key. Set the api_key value in the configuration or use the LWS_API_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if timeout <= 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("timeout"),
			"Invalid timeout value",
			"The timeout value must be a positive integer representing seconds.",
		)
	}

	if retries < 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("retries"),
			"Invalid retries value",
			"The retries value cannot be negative.",
		)
	}

	if delay < 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("delay"),
			"Invalid delay value",
			"The delay value cannot be negative.",
		)
	}

	if backoff < 1 {
		resp.Diagnostics.AddAttributeError(
			path.Root("backoff"),
			"Invalid backoff value",
			"The backoff value must be at least 1.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new LWS client using the configuration values
	lwsClient := client.NewLWSClient(login, apiKey, baseUrl, testMode, timeout, retries, delay, backoff)

	// Make the LWS client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = lwsClient
	resp.ResourceData = lwsClient
}

func (p *LWSProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDNSRecordResource,
	}
}

func (p *LWSProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDNSZoneDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &LWSProvider{
			version: version,
		}
	}
}
