// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/VanGoghDev/gophermart/internal/router (interfaces: Storage)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	models "github.com/VanGoghDev/gophermart/internal/domain/models"
	gomock "github.com/golang/mock/gomock"
)

// MockStorage is a mock of Storage interface.
type MockStorage struct {
	ctrl     *gomock.Controller
	recorder *MockStorageMockRecorder
}

// MockStorageMockRecorder is the mock recorder for MockStorage.
type MockStorageMockRecorder struct {
	mock *MockStorage
}

// NewMockStorage creates a new mock instance.
func NewMockStorage(ctrl *gomock.Controller) *MockStorage {
	mock := &MockStorage{ctrl: ctrl}
	mock.recorder = &MockStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorage) EXPECT() *MockStorageMockRecorder {
	return m.recorder
}

// GetBalance mocks base method.
func (m *MockStorage) GetBalance(arg0 context.Context, arg1 string) (models.Balance, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBalance", arg0, arg1)
	ret0, _ := ret[0].(models.Balance)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBalance indicates an expected call of GetBalance.
func (mr *MockStorageMockRecorder) GetBalance(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBalance", reflect.TypeOf((*MockStorage)(nil).GetBalance), arg0, arg1)
}

// GetOrder mocks base method.
func (m *MockStorage) GetOrder(arg0 context.Context, arg1 string) (models.Order, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrder", arg0, arg1)
	ret0, _ := ret[0].(models.Order)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrder indicates an expected call of GetOrder.
func (mr *MockStorageMockRecorder) GetOrder(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrder", reflect.TypeOf((*MockStorage)(nil).GetOrder), arg0, arg1)
}

// GetOrders mocks base method.
func (m *MockStorage) GetOrders(arg0 context.Context, arg1 string) ([]models.Order, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrders", arg0, arg1)
	ret0, _ := ret[0].([]models.Order)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrders indicates an expected call of GetOrders.
func (mr *MockStorageMockRecorder) GetOrders(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrders", reflect.TypeOf((*MockStorage)(nil).GetOrders), arg0, arg1)
}

// GetUser mocks base method.
func (m *MockStorage) GetUser(arg0 context.Context, arg1, arg2 string) (models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUser", arg0, arg1, arg2)
	ret0, _ := ret[0].(models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUser indicates an expected call of GetUser.
func (mr *MockStorageMockRecorder) GetUser(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUser", reflect.TypeOf((*MockStorage)(nil).GetUser), arg0, arg1, arg2)
}

// GetWithdrawals mocks base method.
func (m *MockStorage) GetWithdrawals(arg0 context.Context, arg1 string) ([]models.Withdrawal, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWithdrawals", arg0, arg1)
	ret0, _ := ret[0].([]models.Withdrawal)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetWithdrawals indicates an expected call of GetWithdrawals.
func (mr *MockStorageMockRecorder) GetWithdrawals(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWithdrawals", reflect.TypeOf((*MockStorage)(nil).GetWithdrawals), arg0, arg1)
}

// RegisterUser mocks base method.
func (m *MockStorage) RegisterUser(arg0 context.Context, arg1, arg2 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RegisterUser", arg0, arg1, arg2)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RegisterUser indicates an expected call of RegisterUser.
func (mr *MockStorageMockRecorder) RegisterUser(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterUser", reflect.TypeOf((*MockStorage)(nil).RegisterUser), arg0, arg1, arg2)
}

// SaveOrder mocks base method.
func (m *MockStorage) SaveOrder(arg0 context.Context, arg1, arg2 string, arg3 models.OrderStatus) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveOrder", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveOrder indicates an expected call of SaveOrder.
func (mr *MockStorageMockRecorder) SaveOrder(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveOrder", reflect.TypeOf((*MockStorage)(nil).SaveOrder), arg0, arg1, arg2, arg3)
}

// SaveWithdrawal mocks base method.
func (m *MockStorage) SaveWithdrawal(arg0 context.Context, arg1, arg2 string, arg3 int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveWithdrawal", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveWithdrawal indicates an expected call of SaveWithdrawal.
func (mr *MockStorageMockRecorder) SaveWithdrawal(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveWithdrawal", reflect.TypeOf((*MockStorage)(nil).SaveWithdrawal), arg0, arg1, arg2, arg3)
}
