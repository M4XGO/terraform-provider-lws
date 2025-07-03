package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/M4XGO/terraform-provider-lws/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &DNSZoneDataSource{}

func NewDNSZoneDataSource() datasource.DataSource {
	return &DNSZoneDataSource{}
}

// DNSZoneDataSource defines the data source implementation.
type DNSZoneDataSource struct {
	client *client.LWSClient
}

// DNSZoneDataSourceModel describes the data source data model.
type DNSZoneDataSourceModel struct {
	Name    types.String         `tfsdk:"name"`
	Records []DNSRecordDataModel `tfsdk:"records"`
}

type DNSRecordDataModel struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
	TTL   types.Int64  `tfsdk:"ttl"`
}

func (d *DNSZoneDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_zone"
}

func (d *DNSZoneDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "LWS DNS zone data source",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "DNS zone name",
				Required:            true,
			},
			"records": schema.ListNestedAttribute{
				MarkdownDescription: "DNS records in the zone",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "DNS record identifier",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "DNS record name",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "DNS record type",
							Computed:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "DNS record value",
							Computed:            true,
						},
						"ttl": schema.Int64Attribute{
							MarkdownDescription: "DNS record TTL",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *DNSZoneDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	lwsClient, ok := req.ProviderData.(*client.LWSClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.LWSClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = lwsClient
}

func (d *DNSZoneDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DNSZoneDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	zoneName := data.Name.ValueString()
	tflog.Info(ctx, "Reading DNS zone", map[string]interface{}{
		"zone_name": zoneName,
		"base_url":  d.client.BaseURL,
		"login":     d.client.Login,
		"test_mode": d.client.TestMode,
	})

	// Get DNS zone information from LWS API
	zone, err := d.client.GetDNSZone(ctx, zoneName)
	if err != nil {
		tflog.Error(ctx, "Failed to read DNS zone", map[string]interface{}{
			"zone_name": zoneName,
			"error":     err.Error(),
			"base_url":  d.client.BaseURL,
			"login":     d.client.Login,
		})

		// Provide more helpful error message
		errorMsg := fmt.Sprintf("Unable to read DNS zone '%s', got error: %s", zoneName, err)
		if d.client.TestMode {
			errorMsg += "\n\nNote: You're in test mode. Make sure your test server is configured correctly."
		} else {
			errorMsg += fmt.Sprintf("\n\nAPI Details:\n- Base URL: %s\n- Login: %s\n- Expected endpoint: %s/v1/domain/%s/zdns",
				d.client.BaseURL, d.client.Login, d.client.BaseURL, zoneName)
		}

		resp.Diagnostics.AddError("Client Error", errorMsg)
		return
	}

	tflog.Debug(ctx, "Successfully retrieved DNS zone", map[string]interface{}{
		"zone_name":    zoneName,
		"record_count": len(zone.Records),
	})

	// Convert DNS records to data model
	records := make([]DNSRecordDataModel, len(zone.Records))
	for i, record := range zone.Records {
		records[i] = DNSRecordDataModel{
			ID:    types.StringValue(strconv.Itoa(record.ID)),
			Name:  types.StringValue(record.Name),
			Type:  types.StringValue(record.Type),
			Value: types.StringValue(record.Value),
			TTL:   types.Int64Value(int64(record.TTL)),
		}
	}

	data.Records = records

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
