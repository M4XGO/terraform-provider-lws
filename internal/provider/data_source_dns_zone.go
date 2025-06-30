package provider

import (
	"context"
	"fmt"

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
	client *LWSClient
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
	Zone  types.String `tfsdk:"zone"`
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
						"zone": schema.StringAttribute{
							MarkdownDescription: "DNS zone name",
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

	client, ok := req.ProviderData.(*LWSClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *LWSClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *DNSZoneDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DNSZoneDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get DNS zone information from LWS API
	zone, err := d.client.GetDNSZone(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read DNS zone, got error: %s", err))
		return
	}

	// Convert DNS records to data model
	records := make([]DNSRecordDataModel, len(zone.Records))
	for i, record := range zone.Records {
		records[i] = DNSRecordDataModel{
			ID:    types.StringValue(record.ID),
			Name:  types.StringValue(record.Name),
			Type:  types.StringValue(record.Type),
			Value: types.StringValue(record.Value),
			TTL:   types.Int64Value(int64(record.TTL)),
			Zone:  types.StringValue(record.Zone),
		}
	}

	data.Records = records

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
