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

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TrafficFilterModel struct {
	ID               types.String             `tfsdk:"id"`
	Name             types.String             `tfsdk:"name"`
	Type             types.String             `tfsdk:"type"`
	Region           types.String             `tfsdk:"region"`
	Description      types.String             `tfsdk:"description"`
	IncludeByDefault types.Bool               `tfsdk:"include_by_default"`
	Rules            []TrafficFilterRuleModel `tfsdk:"rule"`
}

type TrafficFilterRuleModel struct {
	Source      types.String `tfsdk:"source"`
	Description types.String `tfsdk:"description"`
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Provides an Elastic Cloud serverless traffic filter resource, which allows traffic filter rules to be created, updated, and deleted. Traffic filter rules are used to limit inbound traffic to serverless project resources.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of this resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the traffic filter",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "Type of the traffic filter. It can be `ip` or `vpce`",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Description: "Filter region, the traffic filter can only be attached to projects in the specific region",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"include_by_default": schema.BoolAttribute{
				Description: "Indicates that the traffic filter should be automatically included in new projects (Defaults to false)",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"description": schema.StringAttribute{
				Description: "Traffic filter description",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"rule": schema.SetNestedBlock{
				Description: "Set of rules, which the traffic filter is made of.",
				Validators:  []validator.Set{setvalidator.SizeAtLeast(1)},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"source": schema.StringAttribute{
							Description: "Traffic filter source: IP address, CIDR mask, or VPC endpoint ID",
							Required:    true,
						},
						"description": schema.StringAttribute{
							Description: "Description of this individual rule",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func stringValue(s string) types.String {
	return types.StringValue(s)
}

func boolValue(b bool) types.Bool {
	return types.BoolValue(b)
}
