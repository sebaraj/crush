/***************************************************************************
 * File Name: user-service/mocks/mock_opensearch.go
 * Author: Bryan SebaRaj
 * Description: Opensearch mock client for testing
 * Date Created: 01-01-2025
 *
 * Copyright (c) 2025 Bryan SebaRaj. All rights reserved.
 *
 * License:
 * This file is part of Crush. See the LICENSE file for details.
 ***************************************************************************/

package mocks

import (
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"github.com/stretchr/testify/mock"
)

type MockOpenSearchClient struct {
	mock.Mock
}

func (m *MockOpenSearchClient) Search(req *opensearchapi.SearchRequest) (*opensearchapi.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*opensearchapi.Response), args.Error(1)
}
