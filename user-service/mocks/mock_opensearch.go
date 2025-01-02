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
