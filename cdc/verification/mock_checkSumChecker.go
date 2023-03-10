// Code generated by mockery v2.14.1. DO NOT EDIT.

package verification

import (
	context "context"

	filter "github.com/pingcap/tiflow/pkg/filter"
	mock "github.com/stretchr/testify/mock"
)

// mockCheckSumChecker is an autogenerated mock type for the checkSumChecker type
type mockCheckSumChecker struct {
	mock.Mock
}

// getAllDBs provides a mock function with given fields: ctx
func (_m *mockCheckSumChecker) getAllDBs(ctx context.Context) ([]string, error) {
	ret := _m.Called(ctx)

	var r0 []string
	if rf, ok := ret.Get(0).(func(context.Context) []string); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// getCheckSum provides a mock function with given fields: ctx, db, f
func (_m *mockCheckSumChecker) getCheckSum(ctx context.Context, db string, f filter.Filter) (map[string]string, error) {
	ret := _m.Called(ctx, db, f)

	var r0 map[string]string
	if rf, ok := ret.Get(0).(func(context.Context, string, filter.Filter) map[string]string); ok {
		r0 = rf(ctx, db, f)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, filter.Filter) error); ok {
		r1 = rf(ctx, db, f)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTnewMockCheckSumChecker interface {
	mock.TestingT
	Cleanup(func())
}

// newMockCheckSumChecker creates a new instance of mockCheckSumChecker. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func newMockCheckSumChecker(t mockConstructorTestingTnewMockCheckSumChecker) *mockCheckSumChecker {
	mock := &mockCheckSumChecker{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
