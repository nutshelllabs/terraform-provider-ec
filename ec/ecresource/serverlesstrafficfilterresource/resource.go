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

package serverlesstrafficfilterresource

import (
	"context"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-ec/ec/internal"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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
	resp.TypeName = req.ProviderTypeName + "_serverless_traffic_filter"
}

func (r *Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	clients, diags := internal.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = clients.Serverless
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model TrafficFilterModel
	diags := req.Plan.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := serverless.CreateTrafficFilterRequest{
		Name:             model.Name.ValueString(),
		Region:           model.Region.ValueString(),
		Type:             serverless.TrafficFilterType(model.Type.ValueString()),
		Description:      model.Description.ValueStringPointer(),
		IncludeByDefault: model.IncludeByDefault.ValueBoolPointer(),
	}

	if len(model.Rules) > 0 {
		rules := make([]serverless.TrafficFilterRule, 0, len(model.Rules))
		for _, rule := range model.Rules {
			rules = append(rules, serverless.TrafficFilterRule{
				Source:      rule.Source.ValueString(),
				Description: rule.Description.ValueStringPointer(),
			})
		}
		createReq.Rules = &rules
	}

	createResp, err := r.client.CreateTrafficFilterWithResponse(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create traffic filter", err.Error())
		return
	}

	if createResp.JSON201 == nil {
		resp.Diagnostics.AddError(
			"Failed to create traffic filter",
			fmt.Sprintf("The API request failed with: %d %s\n%s",
				createResp.StatusCode(),
				createResp.Status(),
				string(createResp.Body)),
		)
		return
	}

	model = modelFromResponse(createResp.JSON201)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model TrafficFilterModel
	diags := req.State.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	readResp, err := r.client.GetTrafficFilterWithResponse(ctx, model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read traffic filter", err.Error())
		return
	}

	if readResp.HTTPResponse != nil && readResp.HTTPResponse.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if readResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Failed to read traffic filter",
			fmt.Sprintf("The API request failed with: %d %s\n%s",
				readResp.StatusCode(),
				readResp.Status(),
				string(readResp.Body)),
		)
		return
	}

	model = modelFromResponse(readResp.JSON200)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model TrafficFilterModel
	diags := req.Plan.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	patchReq := serverless.PatchTrafficFilterRequest{
		Name:             model.Name.ValueStringPointer(),
		Description:      model.Description.ValueStringPointer(),
		IncludeByDefault: model.IncludeByDefault.ValueBoolPointer(),
	}

	if len(model.Rules) > 0 {
		rules := make([]serverless.TrafficFilterRule, 0, len(model.Rules))
		for _, rule := range model.Rules {
			rules = append(rules, serverless.TrafficFilterRule{
				Source:      rule.Source.ValueString(),
				Description: rule.Description.ValueStringPointer(),
			})
		}
		patchReq.Rules = &rules
	}

	patchResp, err := r.client.PatchTrafficFilterWithResponse(ctx, model.ID.ValueString(), patchReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update traffic filter", err.Error())
		return
	}

	if patchResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Failed to update traffic filter",
			fmt.Sprintf("The API request failed with: %d %s\n%s",
				patchResp.StatusCode(),
				patchResp.Status(),
				string(patchResp.Body)),
		)
		return
	}

	model = modelFromResponse(patchResp.JSON200)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model TrafficFilterModel
	diags := req.State.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteResp, err := r.client.DeleteTrafficFilterWithResponse(ctx, model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete traffic filter", err.Error())
		return
	}

	statusCode := deleteResp.StatusCode()
	if statusCode != http.StatusOK && statusCode != http.StatusNoContent && statusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Failed to delete traffic filter",
			fmt.Sprintf("The API request failed with: %d %s\n%s",
				deleteResp.StatusCode(),
				deleteResp.Status(),
				string(deleteResp.Body)),
		)
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func modelFromResponse(info *serverless.TrafficFilterInfo) TrafficFilterModel {
	model := TrafficFilterModel{}
	model.ID = stringValue(info.Id)
	model.Name = stringValue(info.Name)
	model.Region = stringValue(info.Region)
	model.Type = stringValue(string(info.Type))
	model.IncludeByDefault = boolValue(info.IncludeByDefault)

	if info.Description != nil && *info.Description != "" {
		model.Description = stringValue(*info.Description)
	}

	if len(info.Rules) > 0 {
		model.Rules = make([]TrafficFilterRuleModel, 0, len(info.Rules))
		for _, rule := range info.Rules {
			ruleModel := TrafficFilterRuleModel{
				Source: stringValue(rule.Source),
			}
			if rule.Description != nil && *rule.Description != "" {
				ruleModel.Description = stringValue(*rule.Description)
			}
			model.Rules = append(model.Rules, ruleModel)
		}
	}

	return model
}
