// Code generated by MockGen. DO NOT EDIT.
// Source: interface.go
//
// Generated by this command:
//
//	mockgen -source interface.go -destination ./mock_check_resolver.go -package graph CheckResolver
//

// Package graph is a generated GoMock package.
package graph

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockCheckResolver is a mock of CheckResolver interface.
type MockCheckResolver struct {
	ctrl     *gomock.Controller
	recorder *MockCheckResolverMockRecorder
}

// MockCheckResolverMockRecorder is the mock recorder for MockCheckResolver.
type MockCheckResolverMockRecorder struct {
	mock *MockCheckResolver
}

// NewMockCheckResolver creates a new mock instance.
func NewMockCheckResolver(ctrl *gomock.Controller) *MockCheckResolver {
	mock := &MockCheckResolver{ctrl: ctrl}
	mock.recorder = &MockCheckResolverMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCheckResolver) EXPECT() *MockCheckResolverMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockCheckResolver) Close() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close")
}

// Close indicates an expected call of Close.
func (mr *MockCheckResolverMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockCheckResolver)(nil).Close))
}

// ResolveCheck mocks base method.
func (m *MockCheckResolver) ResolveCheck(ctx context.Context, req *ResolveCheckRequest) (*ResolveCheckResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ResolveCheck", ctx, req)
	ret0, _ := ret[0].(*ResolveCheckResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ResolveCheck indicates an expected call of ResolveCheck.
func (mr *MockCheckResolverMockRecorder) ResolveCheck(ctx, req any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ResolveCheck", reflect.TypeOf((*MockCheckResolver)(nil).ResolveCheck), ctx, req)
}
