package mocks

import (
	"github.com/opensearch-project/opensearch-go"
	"github.com/stretchr/testify/mock"
)

type MockOpenSearchClient struct {
	mock.Mock
}

func (m *MockOpenSearchClient) Search(options ...func(*opensearch.SearchRequest)) (*opensearch.Response, error) {
	args := m.Called(options)
	return args.Get(0).(*opensearch.Response), args.Error(1)
}
