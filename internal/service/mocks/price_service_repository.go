// Code generated by mockery v2.30.1. DO NOT EDIT.

package mocks

import (
	context "context"

	model "github.com/distuurbia/PriceService/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// PriceServiceRepository is an autogenerated mock type for the PriceServiceRepository type
type PriceServiceRepository struct {
	mock.Mock
}

// ReadFromStream provides a mock function with given fields: ctx
func (_m *PriceServiceRepository) ReadFromStream(ctx context.Context) ([]*model.Share, error) {
	ret := _m.Called(ctx)

	var r0 []*model.Share
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]*model.Share, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []*model.Share); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Share)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewPriceServiceRepository creates a new instance of PriceServiceRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewPriceServiceRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *PriceServiceRepository {
	mock := &PriceServiceRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
