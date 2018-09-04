// Code generated by MockGen. DO NOT EDIT.
// Source: agent.go

// Package agent is a generated GoMock package.
package agent

import (
	http "net/http"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	types "github.com/stratumn/go-indigocore/types"
)

// MockAgent is a mock of Agent interface
type MockAgent struct {
	ctrl     *gomock.Controller
	recorder *MockAgentMockRecorder
}

// MockAgentMockRecorder is the mock recorder for MockAgent
type MockAgentMockRecorder struct {
	mock *MockAgent
}

// NewMockAgent creates a new mock instance
func NewMockAgent(ctrl *gomock.Controller) *MockAgent {
	mock := &MockAgent{ctrl: ctrl}
	mock.recorder = &MockAgentMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockAgent) EXPECT() *MockAgentMockRecorder {
	return m.recorder
}

// AddProcess mocks base method
func (m *MockAgent) AddProcess(process string, actions Actions, storeClient interface{}, fossilizerClients []interface{}, opts *ProcessOptions) (Process, error) {
	ret := m.ctrl.Call(m, "AddProcess", process, actions, storeClient, fossilizerClients, opts)
	ret0, _ := ret[0].(Process)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddProcess indicates an expected call of AddProcess
func (mr *MockAgentMockRecorder) AddProcess(process, actions, storeClient, fossilizerClients, opts interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddProcess", reflect.TypeOf((*MockAgent)(nil).AddProcess), process, actions, storeClient, fossilizerClients, opts)
}

// UploadProcess mocks base method
func (m *MockAgent) UploadProcess(processName string, actionsPath string, storeURL string, fossilizerURLs []string, pluginIDs []string) (*Process, error) {
	ret := m.ctrl.Call(m, "UploadProcess", processName, actionsPath, storeURL, fossilizerURLs, pluginIDs)
	ret0, _ := ret[0].(*Process)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UploadProcess indicates an expected call of AddProcess
func (mr *MockAgentMockRecorder) UploadProcess(processName, actionsPath, storeURL, fossilizerURLs, pluginIDs interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UploadProcess", reflect.TypeOf((*MockAgent)(nil).UploadProcess), processName, actionsPath, storeURL, fossilizerURLs, pluginIDs)
}

// FindSegments mocks base method
func (m *MockAgent) FindSegments(process string, opts map[string]string) (types.SegmentSlice, error) {
	ret := m.ctrl.Call(m, "FindSegments", process, opts)
	ret0, _ := ret[0].(types.SegmentSlice)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindSegments indicates an expected call of FindSegments
func (mr *MockAgentMockRecorder) FindSegments(process, opts interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindSegments", reflect.TypeOf((*MockAgent)(nil).FindSegments), process, opts)
}

// GetInfo mocks base method
func (m *MockAgent) GetInfo() (*Info, error) {
	ret := m.ctrl.Call(m, "GetInfo")
	ret0, _ := ret[0].(*Info)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetInfo indicates an expected call of GetInfo
func (mr *MockAgentMockRecorder) GetInfo() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInfo", reflect.TypeOf((*MockAgent)(nil).GetInfo))
}

// GetMapIds mocks base method
func (m *MockAgent) GetMapIds(process string, opts map[string]string) (types.SegmentSlice, error) {
	ret := m.ctrl.Call(m, "GetMapIds", process, opts)
	ret0, _ := ret[0].(types.SegmentSlice)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMapIds indicates an expected call of GetMapIds
func (mr *MockAgentMockRecorder) GetMapIds(process, opts interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMapIds", reflect.TypeOf((*MockAgent)(nil).GetMapIds), process, opts)
}

// GetProcesses mocks base method
func (m *MockAgent) GetProcesses() (Processes, error) {
	ret := m.ctrl.Call(m, "GetProcesses")
	ret0, _ := ret[0].(Processes)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProcesses indicates an expected call of GetProcesses
func (mr *MockAgentMockRecorder) GetProcesses() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProcesses", reflect.TypeOf((*MockAgent)(nil).GetProcesses))
}

// GetProcess mocks base method
func (m *MockAgent) GetProcess(process string) (Process, error) {
	ret := m.ctrl.Call(m, "GetProcess", process)
	ret0, _ := ret[0].(Process)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProcess indicates an expected call of GetProcess
func (mr *MockAgentMockRecorder) GetProcess(process interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProcess", reflect.TypeOf((*MockAgent)(nil).GetProcess), process)
}

// HttpServer mocks base method
func (m *MockAgent) HttpServer() *http.Server {
	ret := m.ctrl.Call(m, "HttpServer")
	ret0, _ := ret[0].(*http.Server)
	return ret0
}

// HttpServer indicates an expected call of HttpServer
func (mr *MockAgentMockRecorder) HttpServer() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HttpServer", reflect.TypeOf((*MockAgent)(nil).HttpServer))
}

// RemoveProcess mocks base method
func (m *MockAgent) RemoveProcess(process string) (Processes, error) {
	ret := m.ctrl.Call(m, "RemoveProcess", process)
	ret0, _ := ret[0].(Processes)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RemoveProcess indicates an expected call of RemoveProcess
func (mr *MockAgentMockRecorder) RemoveProcess(process interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveProcess", reflect.TypeOf((*MockAgent)(nil).RemoveProcess), process)
}

// Url mocks base method
func (m *MockAgent) Url() string {
	ret := m.ctrl.Call(m, "Url")
	ret0, _ := ret[0].(string)
	return ret0
}

// Url indicates an expected call of Url
func (mr *MockAgentMockRecorder) Url() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Url", reflect.TypeOf((*MockAgent)(nil).Url))
}
