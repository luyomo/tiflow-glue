// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/pingcap/tiflow/engine/enginepb (interfaces: BrokerServiceClient)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	enginepb "github.com/pingcap/tiflow/engine/enginepb"
	grpc "google.golang.org/grpc"
)

// MockBrokerServiceClient is a mock of BrokerServiceClient interface.
type MockBrokerServiceClient struct {
	ctrl     *gomock.Controller
	recorder *MockBrokerServiceClientMockRecorder
}

// MockBrokerServiceClientMockRecorder is the mock recorder for MockBrokerServiceClient.
type MockBrokerServiceClientMockRecorder struct {
	mock *MockBrokerServiceClient
}

// NewMockBrokerServiceClient creates a new mock instance.
func NewMockBrokerServiceClient(ctrl *gomock.Controller) *MockBrokerServiceClient {
	mock := &MockBrokerServiceClient{ctrl: ctrl}
	mock.recorder = &MockBrokerServiceClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBrokerServiceClient) EXPECT() *MockBrokerServiceClientMockRecorder {
	return m.recorder
}

// RemoveResource mocks base method.
func (m *MockBrokerServiceClient) RemoveResource(arg0 context.Context, arg1 *enginepb.RemoveLocalResourceRequest, arg2 ...grpc.CallOption) (*enginepb.RemoveLocalResourceResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "RemoveResource", varargs...)
	ret0, _ := ret[0].(*enginepb.RemoveLocalResourceResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RemoveResource indicates an expected call of RemoveResource.
func (mr *MockBrokerServiceClientMockRecorder) RemoveResource(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveResource", reflect.TypeOf((*MockBrokerServiceClient)(nil).RemoveResource), varargs...)
}
