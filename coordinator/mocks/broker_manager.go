package mocks

import "github.com/stretchr/testify/mock"

import "github.com/gtfierro/cs262-project/common"

import "net"

// This is an autogenerated mock type for the BrokerManager type
type BrokerManager struct {
	mock.Mock
}

// BroadcastToBrokers provides a mock function with given fields: message
func (_m *BrokerManager) BroadcastToBrokers(message common.Sendable) {
	_m.Called(message)
}

// ConnectBroker provides a mock function with given fields: brokerInfo, conn
func (_m *BrokerManager) ConnectBroker(brokerInfo *common.BrokerInfo, conn *net.TCPConn) error {
	ret := _m.Called(brokerInfo, conn)

	var r0 error
	if rf, ok := ret.Get(0).(func(*common.BrokerInfo, *net.TCPConn) error); ok {
		r0 = rf(brokerInfo, conn)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetBrokerAddr provides a mock function with given fields: brokerID
func (_m *BrokerManager) GetBrokerAddr(brokerID common.UUID) *string {
	ret := _m.Called(brokerID)

	var r0 *string
	if rf, ok := ret.Get(0).(func(common.UUID) *string); ok {
		r0 = rf(brokerID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*string)
		}
	}

	return r0
}

// GetBrokerInfo provides a mock function with given fields: brokerID
func (_m *BrokerManager) GetBrokerInfo(brokerID common.UUID) *common.BrokerInfo {
	ret := _m.Called(brokerID)

	var r0 *common.BrokerInfo
	if rf, ok := ret.Get(0).(func(common.UUID) *common.BrokerInfo); ok {
		r0 = rf(brokerID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*common.BrokerInfo)
		}
	}

	return r0
}

// GetLiveBroker provides a mock function with given fields:
func (_m *BrokerManager) GetLiveBroker() *common.BrokerInfo {
	ret := _m.Called()

	var r0 *common.BrokerInfo
	if rf, ok := ret.Get(0).(func() *common.BrokerInfo); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*common.BrokerInfo)
		}
	}

	return r0
}

// IsBrokerAlive provides a mock function with given fields: brokerID
func (_m *BrokerManager) IsBrokerAlive(brokerID common.UUID) bool {
	ret := _m.Called(brokerID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(common.UUID) bool); ok {
		r0 = rf(brokerID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// RequestHeartbeat provides a mock function with given fields: brokerID
func (_m *BrokerManager) RequestHeartbeat(brokerID common.UUID) error {
	ret := _m.Called(brokerID)

	var r0 error
	if rf, ok := ret.Get(0).(func(common.UUID) error); ok {
		r0 = rf(brokerID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SendToBroker provides a mock function with given fields: brokerID, message
func (_m *BrokerManager) SendToBroker(brokerID common.UUID, message common.Sendable) error {
	ret := _m.Called(brokerID, message)

	var r0 error
	if rf, ok := ret.Get(0).(func(common.UUID, common.Sendable) error); ok {
		r0 = rf(brokerID, message)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// TerminateBroker provides a mock function with given fields: brokerID
func (_m *BrokerManager) TerminateBroker(brokerID common.UUID) {
	_m.Called(brokerID)
}
