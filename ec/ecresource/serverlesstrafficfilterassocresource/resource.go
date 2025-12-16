// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package serverlesstrafficfilterassocresource

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/terraform-provider-ec/ec/internal"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless"
)

var _ resource.Resource = &Resource{}
var _ resource.ResourceWithConfigure = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}

type Resource struct {
	client serverless.ClientWithResponsesInterface
}

func NewResource() resource.Resource {
	return &Resource{}
}

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serverless_traffic_filter_association"
}

func (r *Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	clients, diags := internal.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = clients.Serverless
}

func resourceReady(r *Resource, dg *diag.Diagnostics) bool {
	if r.client == nil {
		dg.AddError(
			"Unconfigured API Client",
			"Expected configured API client. Please report this issue to the provider developers.",
		)
		return false
	}
	return true
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !resourceReady(r, &resp.Diagnostics) {
		return
	}

	var model modelV0
	diags := req.Plan.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	projectType := model.ProjectType.ValueString()
	trafficFilterID := model.TrafficFilterID.ValueString()

	// Get current traffic filters from the project
	currentFilters, diags := r.getProjectTrafficFilters(ctx, projectID, projectType)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if filter is already associated
	for _, f := range currentFilters {
		if f.Id == trafficFilterID {
			// Already associated, just set state
			model.ID = types.StringValue(fmt.Sprintf("%s-%s", projectID, trafficFilterID))
			resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
			return
		}
	}

	// Add the new filter
	newFilters := append(currentFilters, serverless.TrafficFilter{Id: trafficFilterID})

	// Patch the project with updated filters
	diags = r.patchProjectTrafficFilters(ctx, projectID, projectType, newFilters)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.ID = types.StringValue(fmt.Sprintf("%s-%s", projectID, trafficFilterID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if !resourceReady(r, &resp.Diagnostics) {
		return
	}

	var model modelV0
	diags := req.State.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	projectType := model.ProjectType.ValueString()
	trafficFilterID := model.TrafficFilterID.ValueString()

	// Get current traffic filters from the project
	currentFilters, diags := r.getProjectTrafficFilters(ctx, projectID, projectType)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if the association still exists
	found := false
	for _, f := range currentFilters {
		if f.Id == trafficFilterID {
			found = true
			break
		}
	}

	if !found {
		// Association no longer exists
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All attributes require replacement, so Update should never be called
	resp.Diagnostics.AddError(
		"Update not supported",
		"All attributes of this resource require replacement. This is a bug in the provider.",
	)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !resourceReady(r, &resp.Diagnostics) {
		return
	}

	var model modelV0
	diags := req.State.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	projectType := model.ProjectType.ValueString()
	trafficFilterID := model.TrafficFilterID.ValueString()

	// Get current traffic filters from the project
	currentFilters, diags := r.getProjectTrafficFilters(ctx, projectID, projectType)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Remove the filter from the list
	newFilters := make([]serverless.TrafficFilter, 0, len(currentFilters))
	for _, f := range currentFilters {
		if f.Id != trafficFilterID {
			newFilters = append(newFilters, f)
		}
	}

	// Patch the project with updated filters
	diags = r.patchProjectTrafficFilters(ctx, projectID, projectType, newFilters)
	resp.Diagnostics.Append(diags...)
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Expected format: project_id,project_type,traffic_filter_id
	parts := strings.Split(req.ID, ",")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected format: project_id,project_type,traffic_filter_id. Got: %s", req.ID),
		)
		return
	}

	projectID := parts[0]
	projectType := parts[1]
	trafficFilterID := parts[2]

	// Validate project type
	if projectType != "elasticsearch" && projectType != "observability" && projectType != "security" {
		resp.Diagnostics.AddError(
			"Invalid project type",
			fmt.Sprintf("project_type must be one of: elasticsearch, observability, security. Got: %s", projectType),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), fmt.Sprintf("%s-%s", projectID, trafficFilterID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), projectID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_type"), projectType)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("traffic_filter_id"), trafficFilterID)...)
}

// getProjectTrafficFilters retrieves the current traffic filters for a project
func (r *Resource) getProjectTrafficFilters(ctx context.Context, projectID, projectType string) ([]serverless.TrafficFilter, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch projectType {
	case "elasticsearch":
		resp, err := r.client.GetElasticsearchProjectWithResponse(ctx, projectID)
		if err != nil {
			diags.AddError("Failed to read project", err.Error())
			return nil, diags
		}
		if resp.HTTPResponse != nil && resp.HTTPResponse.StatusCode == http.StatusNotFound {
			diags.AddError("Project not found", fmt.Sprintf("Elasticsearch project %s not found", projectID))
			return nil, diags
		}
		if resp.JSON200 == nil {
			diags.AddError(
				"Failed to read project",
				fmt.Sprintf("The API request failed with: %d %s\n%s", resp.StatusCode(), resp.Status(), string(resp.Body)),
			)
			return nil, diags
		}
		if resp.JSON200.TrafficFilters == nil {
			return []serverless.TrafficFilter{}, nil
		}
		return *resp.JSON200.TrafficFilters, nil

	case "observability":
		resp, err := r.client.GetObservabilityProjectWithResponse(ctx, projectID)
		if err != nil {
			diags.AddError("Failed to read project", err.Error())
			return nil, diags
		}
		if resp.HTTPResponse != nil && resp.HTTPResponse.StatusCode == http.StatusNotFound {
			diags.AddError("Project not found", fmt.Sprintf("Observability project %s not found", projectID))
			return nil, diags
		}
		if resp.JSON200 == nil {
			diags.AddError(
				"Failed to read project",
				fmt.Sprintf("The API request failed with: %d %s\n%s", resp.StatusCode(), resp.Status(), string(resp.Body)),
			)
			return nil, diags
		}
		if resp.JSON200.TrafficFilters == nil {
			return []serverless.TrafficFilter{}, nil
		}
		return *resp.JSON200.TrafficFilters, nil

	case "security":
		resp, err := r.client.GetSecurityProjectWithResponse(ctx, projectID)
		if err != nil {
			diags.AddError("Failed to read project", err.Error())
			return nil, diags
		}
		if resp.HTTPResponse != nil && resp.HTTPResponse.StatusCode == http.StatusNotFound {
			diags.AddError("Project not found", fmt.Sprintf("Security project %s not found", projectID))
			return nil, diags
		}
		if resp.JSON200 == nil {
			diags.AddError(
				"Failed to read project",
				fmt.Sprintf("The API request failed with: %d %s\n%s", resp.StatusCode(), resp.Status(), string(resp.Body)),
			)
			return nil, diags
		}
		if resp.JSON200.TrafficFilters == nil {
			return []serverless.TrafficFilter{}, nil
		}
		return *resp.JSON200.TrafficFilters, nil

	default:
		diags.AddError("Invalid project type", fmt.Sprintf("Unknown project type: %s", projectType))
		return nil, diags
	}
}

// patchProjectTrafficFilters updates the traffic filters for a project
func (r *Resource) patchProjectTrafficFilters(ctx context.Context, projectID, projectType string, filters []serverless.TrafficFilter) diag.Diagnostics {
	var diags diag.Diagnostics
	trafficFilters := serverless.OptionalTrafficFilters(filters)

	switch projectType {
	case "elasticsearch":
		patchReq := serverless.PatchElasticsearchProjectRequest{
			TrafficFilters: &trafficFilters,
		}
		resp, err := r.client.PatchElasticsearchProjectWithResponse(ctx, projectID, nil, patchReq)
		if err != nil {
			diags.AddError("Failed to update project", err.Error())
			return diags
		}
		if resp.JSON200 == nil {
			diags.AddError(
				"Failed to update project",
				fmt.Sprintf("The API request failed with: %d %s\n%s", resp.StatusCode(), resp.Status(), string(resp.Body)),
			)
			return diags
		}

	case "observability":
		patchReq := serverless.PatchObservabilityProjectRequest{
			TrafficFilters: &trafficFilters,
		}
		resp, err := r.client.PatchObservabilityProjectWithResponse(ctx, projectID, nil, patchReq)
		if err != nil {
			diags.AddError("Failed to update project", err.Error())
			return diags
		}
		if resp.JSON200 == nil {
			diags.AddError(
				"Failed to update project",
				fmt.Sprintf("The API request failed with: %d %s\n%s", resp.StatusCode(), resp.Status(), string(resp.Body)),
			)
			return diags
		}

	case "security":
		patchReq := serverless.PatchSecurityProjectRequest{
			TrafficFilters: &trafficFilters,
		}
		resp, err := r.client.PatchSecurityProjectWithResponse(ctx, projectID, nil, patchReq)
		if err != nil {
			diags.AddError("Failed to update project", err.Error())
			return diags
		}
		if resp.JSON200 == nil {
			diags.AddError(
				"Failed to update project",
				fmt.Sprintf("The API request failed with: %d %s\n%s", resp.StatusCode(), resp.Status(), string(resp.Body)),
			)
			return diags
		}

	default:
		diags.AddError("Invalid project type", fmt.Sprintf("Unknown project type: %s", projectType))
	}

	return diags
}
