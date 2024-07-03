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
