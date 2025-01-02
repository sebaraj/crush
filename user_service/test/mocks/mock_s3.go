package mocks

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/mock"
)

type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) GetObjectRequest(input *s3.GetObjectInput) (*s3.Request, *s3.GetObjectOutput) {
	args := m.Called(input)
	return args.Get(0).(*s3.Request), args.Get(1).(*s3.GetObjectOutput)
}
