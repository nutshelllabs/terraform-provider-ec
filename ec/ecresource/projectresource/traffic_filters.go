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

package projectresource

import (
	"context"

	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// trafficFiltersFromModel converts a Terraform Set of traffic filter IDs to the API format
func trafficFiltersFromModel(ctx context.Context, tfSet types.Set) (*serverless.TrafficFilters, diag.Diagnostics) {
	if tfSet.IsNull() || tfSet.IsUnknown() {
		return nil, nil
	}

	var ids []string
	diags := tfSet.ElementsAs(ctx, &ids, false)
	if diags.HasError() {
		return nil, diags
	}

	if len(ids) == 0 {
		return nil, nil
	}

	filters := make(serverless.TrafficFilters, 0, len(ids))
	for _, id := range ids {
		filters = append(filters, serverless.TrafficFilter{Id: id})
	}
	return &filters, nil
}

// trafficFiltersToModel converts API traffic filters to a Terraform Set of strings
func trafficFiltersToModel(ctx context.Context, filters *serverless.TrafficFilters) (types.Set, diag.Diagnostics) {
	if filters == nil || len(*filters) == 0 {
		return types.SetNull(types.StringType), nil
	}

	ids := make([]string, 0, len(*filters))
	for _, f := range *filters {
		ids = append(ids, f.Id)
	}

	return types.SetValueFrom(ctx, types.StringType, ids)
}
