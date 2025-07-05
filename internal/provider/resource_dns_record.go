package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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

	tflog.Info(ctx, "Processing DNS record request", map[string]interface{}{
		"name":     record.Name,
		"type":     record.Type,
		"value":    record.Value,
		"zone":     record.Zone,
		"ttl":      record.TTL,
		"base_url": r.client.BaseURL,
		"login":    r.client.Login,
	})

	// First, check if a record with the same name and type already exists
	tflog.Debug(ctx, "Checking if DNS record already exists", map[string]interface{}{
		"name": record.Name,
		"type": record.Type,
		"zone": record.Zone,
	})

	zone, err := r.client.GetDNSZone(ctx, record.Zone)
	if err != nil {
		tflog.Error(ctx, "Failed to get DNS zone for conflict check", map[string]interface{}{
			"zone":  record.Zone,
			"error": err.Error(),
		})
		// If we can't get the zone, continue with create attempt
	} else {
		// Look for existing record with same name and type
		for _, existingRecord := range zone.Records {
			if existingRecord.Name == record.Name && existingRecord.Type == record.Type {
				tflog.Info(ctx, "Found existing DNS record, will update instead of create", map[string]interface{}{
					"existing_id":    existingRecord.ID,
					"existing_value": existingRecord.Value,
					"new_value":      record.Value,
					"name":           record.Name,
					"type":           record.Type,
					"zone":           record.Zone,
				})

				// Validate that the existing record has a valid ID
				if existingRecord.ID <= 0 {
					tflog.Error(ctx, "Found existing record but ID is invalid", map[string]interface{}{
						"existing_id": existingRecord.ID,
						"name":        record.Name,
						"type":        record.Type,
						"zone":        record.Zone,
					})

					errorMsg := fmt.Sprintf("Found existing DNS record '%s' of type '%s' but it has invalid ID: %d. Cannot update record with invalid ID.",
						record.Name, record.Type, existingRecord.ID)
					resp.Diagnostics.AddError("Invalid Record ID", errorMsg)
					return
				}

				// Update existing record instead of creating
				record.ID = existingRecord.ID
				updatedRecord, err := r.client.UpdateDNSRecord(ctx, record)
				if err != nil {
					errorMsg := fmt.Sprintf("Unable to update existing DNS record '%s' (ID: %d) in zone '%s', got error: %s",
						record.Name, existingRecord.ID, record.Zone, err)
					if r.client.TestMode {
						errorMsg += "\n\nNote: You're in test mode. Make sure your test server is configured correctly."
					} else {
						errorMsg += fmt.Sprintf("\n\nAPI Details:\n- Base URL: %s\n- Login: %s\n- Expected endpoint: %s/domain/%s/zdns",
							r.client.BaseURL, r.client.Login, r.client.BaseURL, record.Zone)
					}

					tflog.Error(ctx, "Failed to update existing DNS record", map[string]interface{}{
						"name":        record.Name,
						"zone":        record.Zone,
						"type":        record.Type,
						"value":       record.Value,
						"existing_id": existingRecord.ID,
						"error":       err.Error(),
					})

					resp.Diagnostics.AddError("Client Error", errorMsg)
					return
				}

				// Validate the updated record ID
				if updatedRecord.ID <= 0 {
					tflog.Error(ctx, "Updated record returned invalid ID", map[string]interface{}{
						"returned_id": updatedRecord.ID,
						"name":        record.Name,
						"type":        record.Type,
						"zone":        record.Zone,
					})

					errorMsg := fmt.Sprintf("Update operation returned invalid ID: %d for record '%s'. This indicates an API problem.",
						updatedRecord.ID, record.Name)
					resp.Diagnostics.AddError("Invalid Updated Record ID", errorMsg)
					return
				}

				tflog.Info(ctx, "Successfully updated existing DNS record", map[string]interface{}{
					"id":        updatedRecord.ID,
					"name":      updatedRecord.Name,
					"type":      updatedRecord.Type,
					"zone":      updatedRecord.Zone,
					"new_value": updatedRecord.Value,
					"action":    "updated_existing",
				})

				// Inform user that we updated an existing record instead of creating
				resp.Diagnostics.AddWarning(
					"Updated Existing DNS Record",
					fmt.Sprintf("Found existing DNS record '%s' of type '%s' in zone '%s' (ID: %d). Updated its value from '%s' to '%s' instead of creating a duplicate.",
						record.Name, record.Type, record.Zone, existingRecord.ID, existingRecord.Value, record.Value),
				)

				// Save updated record data into Terraform state
				data.ID = types.StringValue(fmt.Sprintf("%d", updatedRecord.ID))
				data.TTL = types.Int64Value(int64(updatedRecord.TTL))

				// Save data into Terraform state
				resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
				return
			}
		}
	}

	// No existing record found, proceed with creation
	tflog.Info(ctx, "No existing record found, creating new DNS record", map[string]interface{}{
		"name": record.Name,
		"type": record.Type,
		"zone": record.Zone,
	})

	createdRecord, err := r.client.CreateDNSRecord(ctx, record)
	if err != nil {
		// Provide more helpful error message
		errorMsg := fmt.Sprintf("Unable to create DNS record '%s' in zone '%s', got error: %s", record.Name, record.Zone, err)
		if r.client.TestMode {
			errorMsg += "\n\nNote: You're in test mode. Make sure your test server is configured correctly."
		} else {
			errorMsg += fmt.Sprintf("\n\nAPI Details:\n- Base URL: %s\n- Login: %s\n- Expected endpoint: %s/domain/%s/zdns",
				r.client.BaseURL, r.client.Login, r.client.BaseURL, record.Zone)
		}

		tflog.Error(ctx, "Failed to create DNS record", map[string]interface{}{
			"name":  record.Name,
			"zone":  record.Zone,
			"type":  record.Type,
			"value": record.Value,
			"error": err.Error(),
		})

		resp.Diagnostics.AddError("Client Error", errorMsg)
		return
	}

	// Validate the created record ID
	if createdRecord.ID <= 0 {
		tflog.Error(ctx, "Created record returned invalid ID", map[string]interface{}{
			"returned_id": createdRecord.ID,
			"name":        record.Name,
			"type":        record.Type,
			"zone":        record.Zone,
		})

		errorMsg := fmt.Sprintf("Create operation returned invalid ID: %d for record '%s'. This indicates an API problem.",
			createdRecord.ID, record.Name)
		resp.Diagnostics.AddError("Invalid Created Record ID", errorMsg)
		return
	}

	tflog.Info(ctx, "Successfully created new DNS record", map[string]interface{}{
		"id":     createdRecord.ID,
		"name":   createdRecord.Name,
		"type":   createdRecord.Type,
		"zone":   createdRecord.Zone,
		"action": "created_new",
	})

	// Log successful creation (no warning needed for normal operation)

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
	recordName := data.Name.ValueString()
	recordType := data.Type.ValueString()

	tflog.Info(ctx, "Reading DNS record", map[string]interface{}{
		"record_id": recordID,
		"zone":      zoneName,
		"name":      recordName,
		"type":      recordType,
		"base_url":  r.client.BaseURL,
		"login":     r.client.Login,
	})

	// Check if ID is invalid (0 or empty)
	recordIDInt, err := strconv.Atoi(recordID)
	if err != nil || recordIDInt <= 0 {
		tflog.Warn(ctx, "Invalid record ID in state, attempting to find record by name/type", map[string]interface{}{
			"invalid_id": recordID,
			"zone":       zoneName,
			"name":       recordName,
			"type":       recordType,
		})

		// Try to find the record by name and type in the zone
		zone, err := r.client.GetDNSZone(ctx, zoneName)
		if err != nil {
			tflog.Error(ctx, "Failed to get DNS zone to find record by name/type", map[string]interface{}{
				"zone":  zoneName,
				"error": err.Error(),
			})

			// If we can't get the zone, assume the record is deleted
			resp.State.RemoveResource(ctx)
			return
		}

		// Look for the record by name and type
		var foundRecord *client.DNSRecord
		for _, record := range zone.Records {
			if record.Name == recordName && record.Type == recordType {
				foundRecord = &record
				break
			}
		}

		if foundRecord == nil {
			tflog.Info(ctx, "DNS record not found in zone, marking as deleted", map[string]interface{}{
				"zone": zoneName,
				"name": recordName,
				"type": recordType,
			})

			// Record doesn't exist, remove from state
			resp.State.RemoveResource(ctx)
			return
		}

		tflog.Info(ctx, "Found DNS record by name/type, updating ID in state", map[string]interface{}{
			"zone":     zoneName,
			"name":     recordName,
			"type":     recordType,
			"found_id": foundRecord.ID,
			"old_id":   recordID,
		})

		// Update the model with found record data
		data.ID = types.StringValue(fmt.Sprintf("%d", foundRecord.ID))
		data.Name = types.StringValue(foundRecord.Name)
		data.Type = types.StringValue(foundRecord.Type)
		data.Value = types.StringValue(foundRecord.Value)
		data.TTL = types.Int64Value(int64(foundRecord.TTL))
		// Keep the original zone name from state, not from the record
		// data.Zone = types.StringValue(zoneName)

		// Save corrected data into Terraform state
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	// Normal flow: get record by ID
	record, err := r.client.GetDNSRecord(ctx, zoneName, recordID)
	if err != nil {
		tflog.Error(ctx, "Failed to read DNS record", map[string]interface{}{
			"record_id": recordID,
			"zone":      zoneName,
			"error":     err.Error(),
			"base_url":  r.client.BaseURL,
		})

		// Check if it's a "not found" error, in which case we should remove from state
		if strings.Contains(err.Error(), "not found") {
			tflog.Info(ctx, "DNS record not found, removing from state", map[string]interface{}{
				"record_id": recordID,
				"zone":      zoneName,
			})
			resp.State.RemoveResource(ctx)
			return
		}

		errorMsg := fmt.Sprintf("Unable to read DNS record ID '%s' in zone '%s', got error: %s",
			recordID, zoneName, err)
		errorMsg += fmt.Sprintf("\n\nAPI Details:\n- Base URL: %s\n- Login: %s\n- Expected endpoint: %s/domain/%s/zdns",
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
	// Keep the original zone name from state, not from the record
	// data.Zone = types.StringValue(zoneName)

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

	tflog.Info(ctx, "Updating DNS record", map[string]interface{}{
		"record_id": recordID,
		"name":      record.Name,
		"type":      record.Type,
		"value":     record.Value,
		"zone":      record.Zone,
		"ttl":       record.TTL,
		"base_url":  r.client.BaseURL,
		"login":     r.client.Login,
	})

	updatedRecord, err := r.client.UpdateDNSRecord(ctx, record)
	if err != nil {
		errorMsg := fmt.Sprintf("Unable to update DNS record '%s' (ID: %d) in zone '%s', got error: %s",
			record.Name, recordID, record.Zone, err)
		if r.client.TestMode {
			errorMsg += "\n\nNote: You're in test mode. Make sure your test server is configured correctly."
		} else {
			errorMsg += fmt.Sprintf("\n\nAPI Details:\n- Base URL: %s\n- Login: %s\n- Expected endpoint: %s/domain/%s/zdns",
				r.client.BaseURL, r.client.Login, r.client.BaseURL, record.Zone)
		}

		tflog.Error(ctx, "Failed to update DNS record", map[string]interface{}{
			"record_id": recordID,
			"name":      record.Name,
			"zone":      record.Zone,
			"type":      record.Type,
			"value":     record.Value,
			"error":     err.Error(),
		})

		resp.Diagnostics.AddError("Client Error", errorMsg)
		return
	}

	tflog.Info(ctx, "Successfully updated DNS record", map[string]interface{}{
		"record_id": recordID,
		"name":      updatedRecord.Name,
		"type":      updatedRecord.Type,
		"value":     updatedRecord.Value,
		"zone":      updatedRecord.Zone,
		"ttl":       updatedRecord.TTL,
		"action":    "updated",
	})

	// Update the model with the updated data from API response
	data.Name = types.StringValue(updatedRecord.Name)
	data.Type = types.StringValue(updatedRecord.Type)
	data.Value = types.StringValue(updatedRecord.Value)
	data.TTL = types.Int64Value(int64(updatedRecord.TTL))
	data.Zone = types.StringValue(updatedRecord.Zone)

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

	recordID := data.ID.ValueString()
	zoneName := data.Zone.ValueString()

	tflog.Info(ctx, "Deleting DNS record", map[string]interface{}{
		"record_id": recordID,
		"zone":      zoneName,
		"base_url":  r.client.BaseURL,
		"login":     r.client.Login,
	})

	// Delete API call logic
	err := r.client.DeleteDNSRecord(ctx, recordID, zoneName)
	if err != nil {
		errorMsg := fmt.Sprintf("Unable to delete DNS record ID '%s' in zone '%s', got error: %s",
			recordID, zoneName, err)
		if r.client.TestMode {
			errorMsg += "\n\nNote: You're in test mode. Make sure your test server is configured correctly."
		} else {
			errorMsg += fmt.Sprintf("\n\nAPI Details:\n- Base URL: %s\n- Login: %s\n- Expected endpoint: %s/domain/%s/zdns",
				r.client.BaseURL, r.client.Login, r.client.BaseURL, zoneName)
		}

		tflog.Error(ctx, "Failed to delete DNS record", map[string]interface{}{
			"record_id": recordID,
			"zone":      zoneName,
			"error":     err.Error(),
		})

		resp.Diagnostics.AddError("Client Error", errorMsg)
		return
	}

	tflog.Info(ctx, "Successfully deleted DNS record", map[string]interface{}{
		"record_id": recordID,
		"zone":      zoneName,
		"action":    "deleted",
	})
}

func (r *DNSRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
