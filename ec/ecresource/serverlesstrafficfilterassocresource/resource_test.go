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
	"net/http"
	"testing"

	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless/mocks"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetProjectTrafficFilters_Elasticsearch(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	projectID := "test-project-id"
	filterID := "test-filter-id"

	mockClient := mocks.NewMockClientWithResponsesInterface(ctrl)

	existingFilters := serverless.TrafficFilters{{Id: filterID}}
	getResp := &serverless.GetElasticsearchProjectResponse{
		JSON200: &serverless.ElasticsearchProject{
			Id:             projectID,
			Name:           "test-project",
			TrafficFilters: &existingFilters,
		},
		HTTPResponse: &http.Response{StatusCode: http.StatusOK},
	}
	mockClient.EXPECT().GetElasticsearchProjectWithResponse(ctx, projectID).Return(getResp, nil)

	r := &Resource{client: mockClient}
	filters, diags := r.getProjectTrafficFilters(ctx, projectID, "elasticsearch")

	require.False(t, diags.HasError())
	require.Len(t, filters, 1)
	require.Equal(t, filterID, filters[0].Id)
}

func TestGetProjectTrafficFilters_Observability(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	projectID := "test-project-id"
	filterID := "test-filter-id"

	mockClient := mocks.NewMockClientWithResponsesInterface(ctrl)

	existingFilters := serverless.TrafficFilters{{Id: filterID}}
	getResp := &serverless.GetObservabilityProjectResponse{
		JSON200: &serverless.ObservabilityProject{
			Id:             projectID,
			Name:           "test-project",
			TrafficFilters: &existingFilters,
		},
		HTTPResponse: &http.Response{StatusCode: http.StatusOK},
	}
	mockClient.EXPECT().GetObservabilityProjectWithResponse(ctx, projectID).Return(getResp, nil)

	r := &Resource{client: mockClient}
	filters, diags := r.getProjectTrafficFilters(ctx, projectID, "observability")

	require.False(t, diags.HasError())
	require.Len(t, filters, 1)
	require.Equal(t, filterID, filters[0].Id)
}

func TestGetProjectTrafficFilters_Security(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	projectID := "test-project-id"
	filterID := "test-filter-id"

	mockClient := mocks.NewMockClientWithResponsesInterface(ctrl)

	existingFilters := serverless.TrafficFilters{{Id: filterID}}
	getResp := &serverless.GetSecurityProjectResponse{
		JSON200: &serverless.SecurityProject{
			Id:             projectID,
			Name:           "test-project",
			TrafficFilters: &existingFilters,
		},
		HTTPResponse: &http.Response{StatusCode: http.StatusOK},
	}
	mockClient.EXPECT().GetSecurityProjectWithResponse(ctx, projectID).Return(getResp, nil)

	r := &Resource{client: mockClient}
	filters, diags := r.getProjectTrafficFilters(ctx, projectID, "security")

	require.False(t, diags.HasError())
	require.Len(t, filters, 1)
	require.Equal(t, filterID, filters[0].Id)
}

func TestGetProjectTrafficFilters_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	projectID := "test-project-id"

	mockClient := mocks.NewMockClientWithResponsesInterface(ctrl)

	getResp := &serverless.GetElasticsearchProjectResponse{
		HTTPResponse: &http.Response{StatusCode: http.StatusNotFound},
	}
	mockClient.EXPECT().GetElasticsearchProjectWithResponse(ctx, projectID).Return(getResp, nil)

	r := &Resource{client: mockClient}
	_, diags := r.getProjectTrafficFilters(ctx, projectID, "elasticsearch")

	require.True(t, diags.HasError())
}

func TestGetProjectTrafficFilters_EmptyFilters(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	projectID := "test-project-id"

	mockClient := mocks.NewMockClientWithResponsesInterface(ctrl)

	getResp := &serverless.GetElasticsearchProjectResponse{
		JSON200: &serverless.ElasticsearchProject{
			Id:             projectID,
			Name:           "test-project",
			TrafficFilters: nil,
		},
		HTTPResponse: &http.Response{StatusCode: http.StatusOK},
	}
	mockClient.EXPECT().GetElasticsearchProjectWithResponse(ctx, projectID).Return(getResp, nil)

	r := &Resource{client: mockClient}
	filters, diags := r.getProjectTrafficFilters(ctx, projectID, "elasticsearch")

	require.False(t, diags.HasError())
	require.Len(t, filters, 0)
}

func TestGetProjectTrafficFilters_InvalidProjectType(t *testing.T) {
	ctx := context.Background()

	r := &Resource{client: nil}
	_, diags := r.getProjectTrafficFilters(ctx, "project-id", "invalid")

	require.True(t, diags.HasError())
}

func TestPatchProjectTrafficFilters_Elasticsearch(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	projectID := "test-project-id"
	filterID := "test-filter-id"

	mockClient := mocks.NewMockClientWithResponsesInterface(ctrl)

	patchResp := &serverless.PatchElasticsearchProjectResponse{
		JSON200: &serverless.ElasticsearchProject{
			Id:   projectID,
			Name: "test-project",
		},
		HTTPResponse: &http.Response{StatusCode: http.StatusOK},
	}
	mockClient.EXPECT().PatchElasticsearchProjectWithResponse(
		ctx,
		projectID,
		(*serverless.PatchElasticsearchProjectParams)(nil),
		gomock.Any(),
	).Return(patchResp, nil)

	r := &Resource{client: mockClient}
	filters := []serverless.TrafficFilter{{Id: filterID}}
	diags := r.patchProjectTrafficFilters(ctx, projectID, "elasticsearch", filters)

	require.False(t, diags.HasError())
}

func TestPatchProjectTrafficFilters_Observability(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	projectID := "test-project-id"
	filterID := "test-filter-id"

	mockClient := mocks.NewMockClientWithResponsesInterface(ctrl)

	patchResp := &serverless.PatchObservabilityProjectResponse{
		JSON200: &serverless.ObservabilityProject{
			Id:   projectID,
			Name: "test-project",
		},
		HTTPResponse: &http.Response{StatusCode: http.StatusOK},
	}
	mockClient.EXPECT().PatchObservabilityProjectWithResponse(
		ctx,
		projectID,
		(*serverless.PatchObservabilityProjectParams)(nil),
		gomock.Any(),
	).Return(patchResp, nil)

	r := &Resource{client: mockClient}
	filters := []serverless.TrafficFilter{{Id: filterID}}
	diags := r.patchProjectTrafficFilters(ctx, projectID, "observability", filters)

	require.False(t, diags.HasError())
}

func TestPatchProjectTrafficFilters_Security(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	projectID := "test-project-id"
	filterID := "test-filter-id"

	mockClient := mocks.NewMockClientWithResponsesInterface(ctrl)

	patchResp := &serverless.PatchSecurityProjectResponse{
		JSON200: &serverless.SecurityProject{
			Id:   projectID,
			Name: "test-project",
		},
		HTTPResponse: &http.Response{StatusCode: http.StatusOK},
	}
	mockClient.EXPECT().PatchSecurityProjectWithResponse(
		ctx,
		projectID,
		(*serverless.PatchSecurityProjectParams)(nil),
		gomock.Any(),
	).Return(patchResp, nil)

	r := &Resource{client: mockClient}
	filters := []serverless.TrafficFilter{{Id: filterID}}
	diags := r.patchProjectTrafficFilters(ctx, projectID, "security", filters)

	require.False(t, diags.HasError())
}

func TestPatchProjectTrafficFilters_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	projectID := "test-project-id"

	mockClient := mocks.NewMockClientWithResponsesInterface(ctrl)

	patchResp := &serverless.PatchElasticsearchProjectResponse{
		HTTPResponse: &http.Response{StatusCode: http.StatusBadRequest},
	}
	mockClient.EXPECT().PatchElasticsearchProjectWithResponse(
		ctx,
		projectID,
		(*serverless.PatchElasticsearchProjectParams)(nil),
		gomock.Any(),
	).Return(patchResp, nil)

	r := &Resource{client: mockClient}
	filters := []serverless.TrafficFilter{{Id: "filter-id"}}
	diags := r.patchProjectTrafficFilters(ctx, projectID, "elasticsearch", filters)

	require.True(t, diags.HasError())
}

func TestResourceReady(t *testing.T) {
	t.Run("returns false when client is nil", func(t *testing.T) {
		r := &Resource{client: nil}
		var diags diag.Diagnostics
		ready := resourceReady(r, &diags)

		require.False(t, ready)
		require.True(t, diags.HasError())
	})

	t.Run("returns true when client is set", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockClient := mocks.NewMockClientWithResponsesInterface(ctrl)

		r := &Resource{client: mockClient}
		var diags diag.Diagnostics
		ready := resourceReady(r, &diags)

		require.True(t, ready)
		require.False(t, diags.HasError())
	})
}
