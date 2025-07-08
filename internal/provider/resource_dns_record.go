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

	// Manual validation for required fields
	recordName := strings.TrimSpace(data.Name.ValueString())
	recordType := strings.TrimSpace(data.Type.ValueString())
	recordValue := strings.TrimSpace(data.Value.ValueString())
	zoneName := strings.TrimSpace(data.Zone.ValueString())

	if recordName == "" {
		resp.Diagnostics.AddError("Validation Error", "DNS record name cannot be empty")
		return
	}

	if recordType == "" {
		resp.Diagnostics.AddError("Validation Error", "DNS record type cannot be empty")
		return
	}

	if recordValue == "" {
		resp.Diagnostics.AddError("Validation Error", "DNS record value cannot be empty")
		return
	}

	if zoneName == "" {
		resp.Diagnostics.AddError("Validation Error", "DNS zone name cannot be empty")
		return
	}

	// Create API call logic
	record := &client.DNSRecord{
		Name:  recordName,
		Type:  recordType,
		Value: recordValue,
		Zone:  zoneName,
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
		// Use case-insensitive comparison and trim whitespace for better matching
		targetName := strings.ToLower(strings.TrimSpace(record.Name))
		targetType := strings.ToUpper(strings.TrimSpace(record.Type))

		tflog.Debug(ctx, "Searching for existing records", map[string]interface{}{
			"target_name":   targetName,
			"target_type":   targetType,
			"total_records": len(zone.Records),
		})

		for _, existingRecord := range zone.Records {
			existingName := strings.ToLower(strings.TrimSpace(existingRecord.Name))
			existingType := strings.ToUpper(strings.TrimSpace(existingRecord.Type))

			tflog.Debug(ctx, "Comparing with existing record", map[string]interface{}{
				"existing_name": existingName,
				"existing_type": existingType,
				"existing_id":   existingRecord.ID,
				"matches_name":  existingName == targetName,
				"matches_type":  existingType == targetType,
			})

			if existingName == targetName && existingType == targetType {
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
		// Check if the error indicates the record already exists
		errorMsg := strings.ToLower(err.Error())
		if strings.Contains(errorMsg, "cannot add record") ||
			strings.Contains(errorMsg, "record invalid") ||
			strings.Contains(errorMsg, "already exists") ||
			strings.Contains(errorMsg, "duplicate") {

			tflog.Warn(ctx, "Create failed, likely due to existing record, attempting to find and adopt it", map[string]interface{}{
				"name":  record.Name,
				"type":  record.Type,
				"zone":  record.Zone,
				"error": err.Error(),
			})

			// Try to fetch the zone again and look more thoroughly for the existing record
			zone, zoneErr := r.client.GetDNSZone(ctx, record.Zone)
			if zoneErr != nil {
				tflog.Error(ctx, "Failed to get DNS zone for fallback search", map[string]interface{}{
					"zone":  record.Zone,
					"error": zoneErr.Error(),
				})
			} else {
				// More thorough search - also check with exact string matching
				for _, existingRecord := range zone.Records {
					// Try both normalized and exact matching
					if (strings.ToLower(strings.TrimSpace(existingRecord.Name)) == strings.ToLower(strings.TrimSpace(record.Name)) &&
						strings.ToUpper(strings.TrimSpace(existingRecord.Type)) == strings.ToUpper(strings.TrimSpace(record.Type))) ||
						(existingRecord.Name == record.Name && existingRecord.Type == record.Type) {

						tflog.Info(ctx, "Found existing record during fallback search, adopting it", map[string]interface{}{
							"existing_id":    existingRecord.ID,
							"existing_name":  existingRecord.Name,
							"existing_type":  existingRecord.Type,
							"existing_value": existingRecord.Value,
							"target_name":    record.Name,
							"target_type":    record.Type,
							"target_value":   record.Value,
						})

						// If values are different, update the record
						if existingRecord.Value != record.Value {
							record.ID = existingRecord.ID
							updatedRecord, updateErr := r.client.UpdateDNSRecord(ctx, record)
							if updateErr != nil {
								tflog.Error(ctx, "Failed to update adopted record", map[string]interface{}{
									"error": updateErr.Error(),
								})
								// Continue with adoption even if update fails
								updatedRecord = &existingRecord
							} else {
								tflog.Info(ctx, "Successfully updated adopted record", map[string]interface{}{
									"id":        updatedRecord.ID,
									"old_value": existingRecord.Value,
									"new_value": updatedRecord.Value,
								})
							}
							createdRecord = updatedRecord
						} else {
							// Values are the same, just adopt the existing record
							createdRecord = &existingRecord
						}

						// Save adopted record data into Terraform state
						data.ID = types.StringValue(fmt.Sprintf("%d", createdRecord.ID))
						data.Name = types.StringValue(createdRecord.Name)
						data.Type = types.StringValue(createdRecord.Type)
						data.Value = types.StringValue(createdRecord.Value)
						data.TTL = types.Int64Value(int64(createdRecord.TTL))
						// Keep the original zone from configuration, not from API response
						data.Zone = types.StringValue(zoneName)

						// Add informational warning
						resp.Diagnostics.AddWarning(
							"Adopted Existing DNS Record",
							fmt.Sprintf("Found existing DNS record '%s' of type '%s' in zone '%s' (ID: %d) that matches the desired configuration. Adopted this record instead of creating a duplicate.",
								record.Name, record.Type, record.Zone, createdRecord.ID),
						)

						// Save data into Terraform state
						resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
						return
					}
				}
			}
		}

		// Original error handling if we couldn't find/adopt an existing record
		fullErrorMsg := fmt.Sprintf("Unable to create DNS record '%s' in zone '%s', got error: %s", record.Name, record.Zone, err)
		if r.client.TestMode {
			fullErrorMsg += "\n\nNote: You're in test mode. Make sure your test server is configured correctly."
		} else {
			fullErrorMsg += fmt.Sprintf("\n\nAPI Details:\n- Base URL: %s\n- Login: %s\n- Expected endpoint: %s/domain/%s/zdns",
				r.client.BaseURL, r.client.Login, r.client.BaseURL, record.Zone)
		}

		tflog.Error(ctx, "Failed to create DNS record", map[string]interface{}{
			"name":  record.Name,
			"zone":  record.Zone,
			"type":  record.Type,
			"value": record.Value,
			"error": err.Error(),
		})

		resp.Diagnostics.AddError("Client Error", fullErrorMsg)
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
	data.Name = types.StringValue(createdRecord.Name)
	data.Type = types.StringValue(createdRecord.Type)
	data.Value = types.StringValue(createdRecord.Value)
	data.TTL = types.Int64Value(int64(createdRecord.TTL))
	// Keep the original zone from configuration, not from API response
	data.Zone = types.StringValue(zoneName)

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

	// DEBUG: Log the current state being read
	tflog.Debug(ctx, "🔍 READ: Starting read operation", map[string]interface{}{
		"state_record_id": recordID,
		"state_zone":      zoneName,
		"state_name":      recordName,
		"state_type":      recordType,
		"state_value":     data.Value.ValueString(),
		"state_ttl":       data.TTL.ValueInt64(),
		"id_is_null":      data.ID.IsNull(),
		"id_is_unknown":   data.ID.IsUnknown(),
		"zone_is_null":    data.Zone.IsNull(),
		"zone_is_unknown": data.Zone.IsUnknown(),
	})

	// Check if zone is missing from state (common issue with older provider versions)
	if zoneName == "" {
		tflog.Error(ctx, "🚨 READ: Zone missing from state", map[string]interface{}{
			"record_id": recordID,
			"name":      recordName,
			"type":      recordType,
		})

		errorMsg := fmt.Sprintf("DNS record zone information is missing from Terraform state. This usually happens when upgrading from an older version of the provider.\n\n"+
			"To fix this issue:\n"+
			"1. Remove the resource from state: terraform state rm lws_dns_record.%s\n"+
			"2. Re-import the resource with the zone: terraform import lws_dns_record.%s zone_name:%s\n"+
			"3. Replace 'zone_name' with the actual DNS zone (e.g., example.com)\n\n"+
			"Record Details:\n- Record ID: %s\n- Name: %s\n- Type: %s",
			// We can't get the actual resource name here, so use placeholder
			"RESOURCE_NAME", "RESOURCE_NAME", recordID, recordID, recordName, recordType)

		resp.Diagnostics.AddError("Missing Zone Information", errorMsg)
		return
	}

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
		tflog.Warn(ctx, "🟡 READ: Invalid record ID in state, attempting to find record by name/type", map[string]interface{}{
			"invalid_id":       recordID,
			"conversion_error": err,
			"zone":             zoneName,
			"name":             recordName,
			"type":             recordType,
		})

		// Try to find the record by name and type in the zone
		zone, err := r.client.GetDNSZone(ctx, zoneName)
		if err != nil {
			tflog.Error(ctx, "🚨 READ: Failed to get DNS zone to find record by name/type", map[string]interface{}{
				"zone":  zoneName,
				"error": err.Error(),
			})

			// If we can't get the zone, assume the record is deleted
			tflog.Info(ctx, "🗑️ READ: Removing resource from state due to zone fetch error", map[string]interface{}{
				"zone":   zoneName,
				"reason": "zone_fetch_failed",
			})
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
			tflog.Info(ctx, "🗑️ READ: DNS record not found in zone, marking as deleted", map[string]interface{}{
				"zone": zoneName,
				"name": recordName,
				"type": recordType,
			})

			// Record doesn't exist, remove from state
			resp.State.RemoveResource(ctx)
			return
		}

		tflog.Info(ctx, "✅ READ: Found DNS record by name/type, updating ID in state", map[string]interface{}{
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
		// Keep the original zone name from state
		data.Zone = types.StringValue(zoneName)

		// Save corrected data into Terraform state
		tflog.Debug(ctx, "💾 READ: Saving corrected state after finding by name", map[string]interface{}{
			"corrected_id":    foundRecord.ID,
			"corrected_name":  foundRecord.Name,
			"corrected_type":  foundRecord.Type,
			"corrected_value": foundRecord.Value,
			"corrected_ttl":   foundRecord.TTL,
		})
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	// Normal flow: get record by ID
	tflog.Debug(ctx, "🔍 READ: Normal flow - fetching record by ID", map[string]interface{}{
		"record_id_int": recordIDInt,
		"zone":          zoneName,
	})

	record, err := r.client.GetDNSRecord(ctx, zoneName, recordID)
	if err != nil {
		tflog.Error(ctx, "🚨 READ: Failed to read DNS record by ID, trying fallback search", map[string]interface{}{
			"record_id": recordID,
			"zone":      zoneName,
			"error":     err.Error(),
			"base_url":  r.client.BaseURL,
		})

		// Check if it's a "not found" error - try fallback search by name/type
		errorMsg := strings.ToLower(err.Error())
		if strings.Contains(errorMsg, "not found") || strings.Contains(errorMsg, "record with id") {
			tflog.Warn(ctx, "🔄 READ: Record ID not found, attempting fallback search by name/type", map[string]interface{}{
				"old_record_id": recordID,
				"zone":          zoneName,
				"name":          recordName,
				"type":          recordType,
				"reason":        "id_changed_or_invalid",
			})

			// Try to find the record by name and type in the zone
			zone, err := r.client.GetDNSZone(ctx, zoneName)
			if err != nil {
				tflog.Error(ctx, "🚨 READ: Failed to get DNS zone for fallback search", map[string]interface{}{
					"zone":  zoneName,
					"error": err.Error(),
				})

				// If we can't get the zone, assume the record is deleted
				tflog.Info(ctx, "🗑️ READ: Removing resource from state due to zone fetch error", map[string]interface{}{
					"zone":   zoneName,
					"reason": "fallback_zone_fetch_failed",
				})
				resp.State.RemoveResource(ctx)
				return
			}

			// Look for the record by name and type
			var foundRecord *client.DNSRecord
			for _, rec := range zone.Records {
				if rec.Name == recordName && rec.Type == recordType {
					foundRecord = &rec
					break
				}
			}

			if foundRecord == nil {
				tflog.Info(ctx, "🗑️ READ: DNS record not found by name/type either, removing from state", map[string]interface{}{
					"zone": zoneName,
					"name": recordName,
					"type": recordType,
				})

				// Record doesn't exist, remove from state
				resp.State.RemoveResource(ctx)
				return
			}

			tflog.Info(ctx, "✅ READ: Found DNS record by name/type, updating ID in state", map[string]interface{}{
				"zone":   zoneName,
				"name":   recordName,
				"type":   recordType,
				"old_id": recordID,
				"new_id": foundRecord.ID,
				"reason": "id_drift_detected",
			})

			// Use the found record
			record = foundRecord
		} else {
			// For other errors, still try to remove from state but log it as a warning
			tflog.Warn(ctx, "🟡 READ: Unable to read DNS record, assuming deleted and removing from state", map[string]interface{}{
				"record_id": recordID,
				"zone":      zoneName,
				"error":     err.Error(),
				"reason":    "api_error_assuming_deleted",
			})

			resp.State.RemoveResource(ctx)
			return
		}
	}

	tflog.Debug(ctx, "✅ READ: Successfully read DNS record from API", map[string]interface{}{
		"record_id": recordID,
		"api_id":    record.ID,
		"api_name":  record.Name,
		"api_type":  record.Type,
		"api_value": record.Value,
		"api_zone":  record.Zone,
		"api_ttl":   record.TTL,
	})

	// Update the model with refreshed data
	data.ID = types.StringValue(fmt.Sprintf("%d", record.ID))
	data.Name = types.StringValue(record.Name)
	data.Type = types.StringValue(record.Type)
	data.Value = types.StringValue(record.Value)
	data.TTL = types.Int64Value(int64(record.TTL))
	// Keep the original zone name from state
	data.Zone = types.StringValue(zoneName)

	// DEBUG: Log what we're saving to state
	tflog.Debug(ctx, "💾 READ: Saving updated state", map[string]interface{}{
		"final_id":    fmt.Sprintf("%d", record.ID),
		"final_name":  record.Name,
		"final_type":  record.Type,
		"final_value": record.Value,
		"final_ttl":   record.TTL,
		"final_zone":  zoneName,
	})

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	// DEBUG: Verify if state was saved correctly
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "🚨 READ: Error saving state", map[string]interface{}{
			"errors": resp.Diagnostics.Errors(),
		})
	} else {
		tflog.Debug(ctx, "✅ READ: State saved successfully", map[string]interface{}{
			"operation": "completed",
		})
	}
}

func (r *DNSRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DNSRecordResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	recordID := strings.TrimSpace(data.ID.ValueString())
	recordName := strings.TrimSpace(data.Name.ValueString())
	recordType := strings.TrimSpace(data.Type.ValueString())
	recordValue := strings.TrimSpace(data.Value.ValueString())
	zoneName := strings.TrimSpace(data.Zone.ValueString())

	// Manual validation for required fields
	if recordID == "" {
		resp.Diagnostics.AddError("Validation Error", "DNS record ID cannot be empty for update operation")
		return
	}

	if recordName == "" {
		resp.Diagnostics.AddError("Validation Error", "DNS record name cannot be empty")
		return
	}

	if recordType == "" {
		resp.Diagnostics.AddError("Validation Error", "DNS record type cannot be empty")
		return
	}

	if recordValue == "" {
		resp.Diagnostics.AddError("Validation Error", "DNS record value cannot be empty")
		return
	}

	if zoneName == "" {
		resp.Diagnostics.AddError("Validation Error", "DNS zone name cannot be empty")
		return
	}

	// Convert string ID to int for validation
	recordIDInt, err := strconv.Atoi(recordID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Record ID",
			fmt.Sprintf("Record ID '%s' is not a valid integer: %s", recordID, err))
		return
	}

	if recordIDInt <= 0 {
		resp.Diagnostics.AddError("Invalid Record ID",
			fmt.Sprintf("Record ID must be a positive integer, got: %d", recordIDInt))
		return
	}

	// Create record object for API call
	record := &client.DNSRecord{
		ID:    recordIDInt,
		Name:  recordName,
		Type:  recordType,
		Value: recordValue,
		Zone:  zoneName,
	}

	if !data.TTL.IsNull() {
		record.TTL = int(data.TTL.ValueInt64())
	}

	tflog.Info(ctx, "Updating DNS record", map[string]interface{}{
		"record_id": recordIDInt,
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
			record.Name, recordIDInt, record.Zone, err)
		if r.client.TestMode {
			errorMsg += "\n\nNote: You're in test mode. Make sure your test server is configured correctly."
		} else {
			errorMsg += fmt.Sprintf("\n\nAPI Details:\n- Base URL: %s\n- Login: %s\n- Expected endpoint: %s/domain/%s/zdns",
				r.client.BaseURL, r.client.Login, r.client.BaseURL, record.Zone)
		}

		tflog.Error(ctx, "Failed to update DNS record", map[string]interface{}{
			"record_id": recordIDInt,
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
		"record_id": recordIDInt,
		"name":      updatedRecord.Name,
		"type":      updatedRecord.Type,
		"value":     updatedRecord.Value,
		"zone":      updatedRecord.Zone,
		"ttl":       updatedRecord.TTL,
		"action":    "updated",
	})

	// Update the model with the updated data from API response
	data.ID = types.StringValue(fmt.Sprintf("%d", updatedRecord.ID))
	data.Name = types.StringValue(updatedRecord.Name)
	data.Type = types.StringValue(updatedRecord.Type)
	data.Value = types.StringValue(updatedRecord.Value)
	data.TTL = types.Int64Value(int64(updatedRecord.TTL))
	// Keep the original zone from configuration, not from API response
	data.Zone = types.StringValue(zoneName)

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

	recordID := strings.TrimSpace(data.ID.ValueString())
	recordName := strings.TrimSpace(data.Name.ValueString())
	recordType := strings.TrimSpace(data.Type.ValueString())
	zoneName := strings.TrimSpace(data.Zone.ValueString())

	// DEBUG: Log the current state being deleted
	tflog.Debug(ctx, "🗑️ DELETE: Starting delete operation", map[string]interface{}{
		"state_record_id": recordID,
		"state_zone":      zoneName,
		"state_name":      recordName,
		"state_type":      recordType,
		"state_value":     data.Value.ValueString(),
		"state_ttl":       data.TTL.ValueInt64(),
		"id_is_null":      data.ID.IsNull(),
		"id_is_unknown":   data.ID.IsUnknown(),
		"zone_is_null":    data.Zone.IsNull(),
		"zone_is_unknown": data.Zone.IsUnknown(),
		"base_url":        r.client.BaseURL,
		"login":           r.client.Login,
		"test_mode":       r.client.TestMode,
	})

	// Manual validation for required fields
	if recordID == "" {
		resp.Diagnostics.AddError("Validation Error", "DNS record ID cannot be empty for delete operation")
		return
	}

	if zoneName == "" {
		resp.Diagnostics.AddError("Validation Error", "DNS zone name cannot be empty for delete operation")
		return
	}

	// Convert string ID to int
	recordIDInt, err := strconv.Atoi(recordID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Record ID",
			fmt.Sprintf("Record ID '%s' is not a valid integer: %s", recordID, err))
		return
	}

	if recordIDInt <= 0 {
		resp.Diagnostics.AddError("Invalid Record ID",
			fmt.Sprintf("Record ID must be a positive integer, got: %d", recordIDInt))
		return
	}

	tflog.Info(ctx, "Deleting DNS record", map[string]interface{}{
		"record_id":   recordIDInt,
		"record_name": recordName,
		"record_type": recordType,
		"zone":        zoneName,
		"base_url":    r.client.BaseURL,
		"login":       r.client.Login,
	})

	// Debug: Log the exact parameters being passed to the API
	tflog.Debug(ctx, "Delete API call parameters", map[string]interface{}{
		"record_id_int": recordIDInt,
		"zone_name":     zoneName,
		"endpoint":      fmt.Sprintf("%s/domain/%s/zdns", r.client.BaseURL, zoneName),
	})

	// Delete API call logic - using ID from state
	err = r.client.DeleteDNSRecord(ctx, recordIDInt, zoneName)
	if err != nil {
		errorMsg := fmt.Sprintf("Unable to delete DNS record ID %d ('%s' of type '%s') in zone '%s', got error: %s",
			recordIDInt, recordName, recordType, zoneName, err)
		if r.client.TestMode {
			errorMsg += "\n\nNote: You're in test mode. Make sure your test server is configured correctly."
		} else {
			errorMsg += fmt.Sprintf("\n\nAPI Details:\n- Base URL: %s\n- Login: %s\n- Expected endpoint: %s/domain/%s/zdns",
				r.client.BaseURL, r.client.Login, r.client.BaseURL, zoneName)
		}

		tflog.Error(ctx, "Failed to delete DNS record", map[string]interface{}{
			"record_id":   recordIDInt,
			"record_name": recordName,
			"record_type": recordType,
			"zone":        zoneName,
			"error":       err.Error(),
		})

		resp.Diagnostics.AddError("Client Error", errorMsg)
		return
	}

	tflog.Info(ctx, "Successfully deleted DNS record", map[string]interface{}{
		"record_id":   recordIDInt,
		"record_name": recordName,
		"record_type": recordType,
		"zone":        zoneName,
		"action":      "deleted",
	})

	// DEBUG: Confirm deletion completion
	tflog.Debug(ctx, "✅ DELETE: Delete operation completed successfully", map[string]interface{}{
		"record_id":         recordIDInt,
		"zone":              zoneName,
		"operation":         "completed",
		"framework_handles": "state_removal",
	})

	// The resource is automatically removed from state by the framework
	// No need to manually clear the state
}

func (r *DNSRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Support two import formats:
	// 1. "record_id" (legacy format, for backward compatibility)
	// 2. "zone:record_id" (new format that includes zone information)

	importID := req.ID

	// Check if the import ID contains a colon (new format)
	if strings.Contains(importID, ":") {
		parts := strings.SplitN(importID, ":", 2)
		if len(parts) != 2 {
			resp.Diagnostics.AddError(
				"Invalid Import ID Format",
				fmt.Sprintf("Expected format 'zone:record_id', got '%s'. Examples:\n"+
					"- terraform import lws_dns_record.example example.com:12345\n"+
					"- terraform import lws_dns_record.example 12345 (legacy format)",
					importID),
			)
			return
		}

		zoneName := strings.TrimSpace(parts[0])
		recordID := strings.TrimSpace(parts[1])

		if zoneName == "" || recordID == "" {
			resp.Diagnostics.AddError(
				"Invalid Import ID Format",
				fmt.Sprintf("Zone and record ID cannot be empty. Got zone='%s', record_id='%s'", zoneName, recordID),
			)
			return
		}

		// Validate record ID is numeric
		if _, err := strconv.Atoi(recordID); err != nil {
			resp.Diagnostics.AddError(
				"Invalid Record ID",
				fmt.Sprintf("Record ID must be a number, got '%s'", recordID),
			)
			return
		}

		tflog.Info(ctx, "Importing DNS record with zone information", map[string]interface{}{
			"zone":      zoneName,
			"record_id": recordID,
			"format":    "zone:record_id",
		})

		// Set both ID and zone in the state
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), recordID)...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("zone"), zoneName)...)

	} else {
		// Legacy format: just the record ID
		// Validate record ID is numeric
		if _, err := strconv.Atoi(importID); err != nil {
			resp.Diagnostics.AddError(
				"Invalid Record ID",
				fmt.Sprintf("Record ID must be a number, got '%s'. For better import experience, use format 'zone:record_id'", importID),
			)
			return
		}

		tflog.Warn(ctx, "Importing DNS record without zone information (legacy format)", map[string]interface{}{
			"record_id":      importID,
			"format":         "record_id_only",
			"recommendation": "Use 'zone:record_id' format for better import experience",
		})

		// Set only the ID, zone will need to be provided manually in the configuration
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), importID)...)

		// Add a warning about the missing zone
		resp.Diagnostics.AddWarning(
			"Zone Information Missing",
			fmt.Sprintf("Imported record ID '%s' without zone information. "+
				"You must specify the 'zone' attribute in your Terraform configuration. "+
				"For a better import experience, use: terraform import lws_dns_record.name zone_name:%s",
				importID, importID),
		)
	}
}
