package service_test

import (
	"context"
	"encoding/json"
	"testing"

	mockclient "github.com/sourcegraph/src-cli/internal/api/mock"
	"github.com/sourcegraph/src-cli/internal/batches/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestService_UpsertBatchChange(t *testing.T) {
	client := new(mockclient.Client)
	mockRequest := new(mockclient.Request)
	svc := service.New(&service.Opts{Client: client})

	tests := []struct {
		name string

		mockInvokes func()

		requestName        string
		requestNamespaceID string

		expectedID   string
		expectedName string
		expectedErr  error
	}{
		{
			name: "New Batch Change",
			mockInvokes: func() {
				client.On("NewRequest", mock.Anything, map[string]interface{}{
					"name":      "my-change",
					"namespace": "my-namespace",
				}).
					Return(mockRequest, nil).
					Once()
				mockRequest.On("Do", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						json.Unmarshal([]byte(`{"upsertEmptyBatchChange":{"id":"123", "name":"my-change"}}`), &args[1])
					}).
					Return(true, nil).
					Once()
			},
			requestName:        "my-change",
			requestNamespaceID: "my-namespace",
			expectedID:         "123",
			expectedName:       "my-change",
		},
		{
			name: "Failed to upsert batch change",
			mockInvokes: func() {
				client.On("NewRequest", mock.Anything, map[string]interface{}{
					"name":      "my-change",
					"namespace": "my-namespace",
				}).
					Return(mockRequest, nil).
					Once()
				mockRequest.On("Do", mock.Anything, mock.Anything).
					Return(false, errors.New("did not get a good response code")).
					Once()
			},
			requestName:        "my-change",
			requestNamespaceID: "my-namespace",
			expectedErr:        errors.New("did not get a good response code"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.mockInvokes != nil {
				test.mockInvokes()
			}

			id, name, err := svc.UpsertBatchChange(context.Background(), test.requestName, test.requestNamespaceID)
			assert.Equal(t, test.expectedID, id)
			assert.Equal(t, test.expectedName, name)
			if test.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			client.AssertExpectations(t)
		})
	}
}

func TestService_CreateBatchSpecFromRaw(t *testing.T) {
	client := new(mockclient.Client)
	mockRequest := new(mockclient.Request)
	svc := service.New(&service.Opts{Client: client})

	tests := []struct {
		name string

		mockInvokes func()

		requestBatchSpec        string
		requestNamespaceID      string
		requestAllowIgnored     bool
		requestAllowUnsupported bool
		requestNoCache          bool
		requestBatchChange      string

		expectedID  string
		expectedErr error
	}{
		{
			name: "Create batch spec",
			mockInvokes: func() {
				client.On("NewRequest", mock.Anything, map[string]interface{}{
					"batchSpec":        "abc",
					"namespace":        "some-namespace",
					"allowIgnored":     false,
					"allowUnsupported": false,
					"noCache":          false,
					"batchChange":      "123",
				}).
					Return(mockRequest, nil).
					Once()
				mockRequest.On("Do", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						json.Unmarshal([]byte(`{"createBatchSpecFromRaw":{"id":"xyz"}}`), &args[1])
					}).
					Return(true, nil).
					Once()
			},
			requestBatchSpec:        "abc",
			requestNamespaceID:      "some-namespace",
			requestAllowIgnored:     false,
			requestAllowUnsupported: false,
			requestNoCache:          false,
			requestBatchChange:      "123",
			expectedID:              "xyz",
		},
		{
			name: "Failed to create batch spec",
			mockInvokes: func() {
				client.On("NewRequest", mock.Anything, map[string]interface{}{
					"batchSpec":        "abc",
					"namespace":        "some-namespace",
					"allowIgnored":     false,
					"allowUnsupported": false,
					"noCache":          false,
					"batchChange":      "123",
				}).
					Return(mockRequest, nil).
					Once()
				mockRequest.On("Do", mock.Anything, mock.Anything).
					Return(false, errors.New("did not get a good response code")).
					Once()
			},
			requestBatchSpec:        "abc",
			requestNamespaceID:      "some-namespace",
			requestAllowIgnored:     false,
			requestAllowUnsupported: false,
			requestNoCache:          false,
			requestBatchChange:      "123",
			expectedErr:             errors.New("did not get a good response code"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.mockInvokes != nil {
				test.mockInvokes()
			}

			id, err := svc.CreateBatchSpecFromRaw(
				context.Background(),
				test.requestBatchSpec,
				test.requestNamespaceID,
				test.requestAllowIgnored,
				test.requestAllowUnsupported,
				test.requestNoCache,
				test.requestBatchChange,
			)
			assert.Equal(t, test.expectedID, id)
			if test.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			client.AssertExpectations(t)
		})
	}
}
