package mocks

import (
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/mock"
	"time"
)

type HeartbeatServiceMock struct {
	mock.Mock
}

func (m *HeartbeatServiceMock) Insert(heartbeat *models.Heartbeat) error {
	args := m.Called(heartbeat)
	return args.Error(0)
}

func (m *HeartbeatServiceMock) InsertBatch(heartbeats []*models.Heartbeat) error {
	args := m.Called(heartbeats)
	return args.Error(0)
}

func (m *HeartbeatServiceMock) Count() (int64, error) {
	args := m.Called()
	return int64(args.Int(0)), args.Error(1)
}

func (m *HeartbeatServiceMock) CountByUser(user *models.User) (int64, error) {
	args := m.Called(user)
	return args.Get(0).(int64), args.Error(0)
}

func (m *HeartbeatServiceMock) CountByUsers(users []*models.User) ([]*models.CountByUser, error) {
	args := m.Called(users)
	return args.Get(0).([]*models.CountByUser), args.Error(0)
}

func (m *HeartbeatServiceMock) GetAllWithin(time time.Time, time2 time.Time, user *models.User) ([]*models.Heartbeat, error) {
	args := m.Called(time, time2, user)
	return args.Get(0).([]*models.Heartbeat), args.Error(1)
}

func (m *HeartbeatServiceMock) GetFirstByUsers() ([]*models.TimeByUser, error) {
	args := m.Called()
	return args.Get(0).([]*models.TimeByUser), args.Error(1)
}

func (m *HeartbeatServiceMock) GetLatestByUser(user *models.User) (*models.Heartbeat, error) {
	args := m.Called(user)
	return args.Get(0).(*models.Heartbeat), args.Error(1)
}

func (m *HeartbeatServiceMock) GetLatestByOriginAndUser(s string, user *models.User) (*models.Heartbeat, error) {
	args := m.Called(s, user)
	return args.Get(0).(*models.Heartbeat), args.Error(1)
}

func (m *HeartbeatServiceMock) GetEntitySetByUser(u uint8, user *models.User) ([]string, error) {
	args := m.Called(u, user)
	return args.Get(0).([]string), args.Error(1)
}

func (m *HeartbeatServiceMock) DeleteBefore(time time.Time) error {
	args := m.Called(time)
	return args.Error(0)
}
