package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &tinymonProvider{}

type TinyMonClient struct {
	URL    string
	APIKey string
	HTTP   *http.Client
}

func (c *TinyMonClient) DoJSON(method, path string, body interface{}, result interface{}) error {
	url := strings.TrimRight(c.URL, "/") + path

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshalling request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("executing request %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API %s %s returned status %d: %s", method, path, resp.StatusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshalling response: %w", err)
		}
	}

	return nil
}

type tinymonProvider struct {
	version string
}

type tinymonProviderModel struct {
	URL    types.String `tfsdk:"url"`
	APIKey types.String `tfsdk:"api_key"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &tinymonProvider{
			version: version,
		}
	}
}

func (p *tinymonProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "tinymon"
	resp.Version = p.version
}

func (p *tinymonProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage TinyMon hosts and checks via the Push API.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Description: "Base URL of the TinyMon instance. Can also be set via TINYMON_URL environment variable.",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "API key (Bearer token) for the Push API. Can also be set via TINYMON_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *tinymonProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config tinymonProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := os.Getenv("TINYMON_URL")
	if !config.URL.IsNull() && !config.URL.IsUnknown() {
		url = config.URL.ValueString()
	}
	if url == "" {
		resp.Diagnostics.AddError(
			"Missing TinyMon URL",
			"Set url in the provider configuration or via the TINYMON_URL environment variable.",
		)
	}

	apiKey := os.Getenv("TINYMON_API_KEY")
	if !config.APIKey.IsNull() && !config.APIKey.IsUnknown() {
		apiKey = config.APIKey.ValueString()
	}
	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing TinyMon API Key",
			"Set api_key in the provider configuration or via the TINYMON_API_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client := &TinyMonClient{
		URL:    url,
		APIKey: apiKey,
		HTTP:   &http.Client{},
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *tinymonProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewHostResource,
		NewCheckResource,
	}
}

func (p *tinymonProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
