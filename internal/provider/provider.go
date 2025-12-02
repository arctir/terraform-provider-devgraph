package provider

import (
	"context"
	"net/http"
	"os"

	v1 "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/oauth2"
)

var _ provider.Provider = &DevgraphProvider{}
var _ v1.SecuritySource = &devgraphSecuritySource{}

type DevgraphProvider struct {
	version string
}

type DevgraphProviderModel struct {
	Host        types.String `tfsdk:"host"`
	AccessToken types.String `tfsdk:"access_token"`
	Environment types.String `tfsdk:"environment"`
}

type devgraphSecuritySource struct {
	token string
}

func (s *devgraphSecuritySource) OAuth2PasswordBearer(ctx context.Context, operationName v1.OperationName) (v1.OAuth2PasswordBearer, error) {
	return v1.OAuth2PasswordBearer{
		Token: s.token,
	}, nil
}

// environmentTransport wraps an http.RoundTripper to add the Devgraph-Environment header
type environmentTransport struct {
	base        http.RoundTripper
	environment string
}

func (t *environmentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.environment != "" {
		req.Header.Set("Devgraph-Environment", t.environment)
	}
	return t.base.RoundTrip(req)
}

func (p *DevgraphProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "devgraph"
	resp.Version = p.version
}

func (p *DevgraphProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for managing Devgraph resources",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "Devgraph API host URL. Can also be set via DEVGRAPH_HOST environment variable.",
				Optional:    true,
			},
			"access_token": schema.StringAttribute{
				Description: "Devgraph API access token. Can also be set via DEVGRAPH_ACCESS_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"environment": schema.StringAttribute{
				Description: "Devgraph environment (organization slug). Can also be set via DEVGRAPH_ENVIRONMENT environment variable.",
				Optional:    true,
			},
		},
	}
}

func (p *DevgraphProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config DevgraphProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Check environment variables if not set in config
	host := os.Getenv("DEVGRAPH_HOST")
	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	accessToken := os.Getenv("DEVGRAPH_ACCESS_TOKEN")
	if !config.AccessToken.IsNull() {
		accessToken = config.AccessToken.ValueString()
	}

	environment := os.Getenv("DEVGRAPH_ENVIRONMENT")
	if !config.Environment.IsNull() {
		environment = config.Environment.ValueString()
	}

	// Validate required fields
	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Devgraph API Host",
			"The provider cannot create the Devgraph API client as there is a missing or empty value for the host. "+
				"Set the host value in the configuration or use the DEVGRAPH_HOST environment variable. ",
		)
	}

	if accessToken == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("access_token"),
			"Missing Devgraph Access Token",
			"The provider cannot create the Devgraph API client as there is a missing or empty value for the access token. "+
				"Set the access_token value in the configuration or use the DEVGRAPH_ACCESS_TOKEN environment variable. ",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create OAuth2 token
	token := &oauth2.Token{
		AccessToken: accessToken,
		TokenType:   "Bearer",
	}

	// Create OAuth2 HTTP client
	oauthConfig := &oauth2.Config{}
	httpClient := oauthConfig.Client(ctx, token)

	// Wrap the HTTP client's transport to add Devgraph-Environment header
	if environment != "" {
		httpClient.Transport = &environmentTransport{
			base:        httpClient.Transport,
			environment: environment,
		}
	}

	// Create security source
	securitySource := &devgraphSecuritySource{token: accessToken}

	// Create Devgraph API client
	client, err := v1.NewClient(host, securitySource, v1.WithClient(httpClient))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Devgraph API Client",
			"An unexpected error occurred when creating the Devgraph API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Devgraph Client Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *DevgraphProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewEnvironmentResource,
		NewMCPEndpointResource,
		NewModelProviderResource,
		NewModelResource,
		NewOAuthServiceResource,
		NewDiscoveryProviderResource,
		NewChatSuggestionResource,
	}
}

func (p *DevgraphProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &DevgraphProvider{
			version: version,
		}
	}
}
