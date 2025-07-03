package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/M4XGO/terraform-provider-lws/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DNSRecordResource{}
var _ resource.ResourceWithImportState = &DNSRecordResource{}

func NewDNSRecordResource() resource.Resource {
	return &DNSRecordResource{}
}

// DNSRecordResource defines the resource implementation.
type DNSRecordResource struct {
	client *client.LWSClient
}

// DNSRecordResourceModel describes the resource data model.
type DNSRecordResourceModel struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
	TTL   types.Int64  `tfsdk:"ttl"`
	Zone  types.String `tfsdk:"zone"`
}

func (r *DNSRecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_record"
}

func (r *DNSRecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "LWS DNS record resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "DNS record identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "DNS record name",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "DNS record type (A, AAAA, CNAME, MX, TXT, etc.)",
				Required:            true,
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "DNS record value",
				Required:            true,
			},
			"ttl": schema.Int64Attribute{
				MarkdownDescription: "DNS record TTL in seconds",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"zone": schema.StringAttribute{
				MarkdownDescription: "DNS zone name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *DNSRecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	lwsClient, ok := req.ProviderData.(*client.LWSClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.LWSClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = lwsClient
}

func (r *DNSRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DNSRecordResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create API call logic
	record := &client.DNSRecord{
		Name:  data.Name.ValueString(),
		Type:  data.Type.ValueString(),
		Value: data.Value.ValueString(),
		Zone:  data.Zone.ValueString(),
	}

	if !data.TTL.IsNull() {
		record.TTL = int(data.TTL.ValueInt64())
	}

	tflog.Info(ctx, "Creating DNS record", map[string]interface{}{
		"name":     record.Name,
		"type":     record.Type,
		"value":    record.Value,
		"zone":     record.Zone,
		"ttl":      record.TTL,
		"base_url": r.client.BaseURL,
		"login":    r.client.Login,
	})

	createdRecord, err := r.client.CreateDNSRecord(ctx, record)
	if err != nil {
		tflog.Error(ctx, "Failed to create DNS record", map[string]interface{}{
			"name":     record.Name,
			"type":     record.Type,
			"zone":     record.Zone,
			"error":    err.Error(),
			"base_url": r.client.BaseURL,
		})

		errorMsg := fmt.Sprintf("Unable to create DNS record '%s' in zone '%s', got error: %s",
			record.Name, record.Zone, err)
		errorMsg += fmt.Sprintf("\n\nAPI Details:\n- Base URL: %s\n- Login: %s\n- Expected endpoint: %s/v1/domain/%s/zdns",
			r.client.BaseURL, r.client.Login, r.client.BaseURL, record.Zone)

		resp.Diagnostics.AddError("Client Error", errorMsg)
		return
	}

	tflog.Debug(ctx, "Successfully created DNS record", map[string]interface{}{
		"id":   createdRecord.ID,
		"name": createdRecord.Name,
		"type": createdRecord.Type,
		"zone": createdRecord.Zone,
	})

	// Save created record data into Terraform state
	data.ID = types.StringValue(fmt.Sprintf("%d", createdRecord.ID))
	data.TTL = types.Int64Value(int64(createdRecord.TTL))

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DNSRecordResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	recordID := data.ID.ValueString()
	zoneName := data.Zone.ValueString()

	tflog.Info(ctx, "Reading DNS record", map[string]interface{}{
		"record_id": recordID,
		"zone":      zoneName,
		"base_url":  r.client.BaseURL,
		"login":     r.client.Login,
	})

	// Get refreshed record value from LWS
	record, err := r.client.GetDNSRecord(ctx, zoneName, recordID)
	if err != nil {
		tflog.Error(ctx, "Failed to read DNS record", map[string]interface{}{
			"record_id": recordID,
			"zone":      zoneName,
			"error":     err.Error(),
			"base_url":  r.client.BaseURL,
		})

		errorMsg := fmt.Sprintf("Unable to read DNS record ID '%s' in zone '%s', got error: %s",
			recordID, zoneName, err)
		errorMsg += fmt.Sprintf("\n\nAPI Details:\n- Base URL: %s\n- Login: %s\n- Expected endpoint: %s/v1/domain/%s/zdns",
			r.client.BaseURL, r.client.Login, r.client.BaseURL, zoneName)

		resp.Diagnostics.AddError("Client Error", errorMsg)
		return
	}

	tflog.Debug(ctx, "Successfully read DNS record", map[string]interface{}{
		"record_id": recordID,
		"name":      record.Name,
		"type":      record.Type,
		"value":     record.Value,
		"zone":      record.Zone,
	})

	// Update the model with refreshed data
	data.Name = types.StringValue(record.Name)
	data.Type = types.StringValue(record.Type)
	data.Value = types.StringValue(record.Value)
	data.TTL = types.Int64Value(int64(record.TTL))
	data.Zone = types.StringValue(record.Zone)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DNSRecordResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert string ID to int
	recordID, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to convert record ID to integer: %s", err))
		return
	}

	// Update API call logic
	record := &client.DNSRecord{
		ID:    recordID,
		Name:  data.Name.ValueString(),
		Type:  data.Type.ValueString(),
		Value: data.Value.ValueString(),
		Zone:  data.Zone.ValueString(),
	}

	if !data.TTL.IsNull() {
		record.TTL = int(data.TTL.ValueInt64())
	}

	updatedRecord, err := r.client.UpdateDNSRecord(ctx, record)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update DNS record, got error: %s", err))
		return
	}

	// Update the model with the updated data
	data.TTL = types.Int64Value(int64(updatedRecord.TTL))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DNSRecordResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete API call logic
	err := r.client.DeleteDNSRecord(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete DNS record, got error: %s", err))
		return
	}
}

func (r *DNSRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
