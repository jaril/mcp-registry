package storage

import (
	"registry/internal/models"

	"github.com/stretchr/testify/mock"
)

// MockStore is a mock implementation of ServerStore for testing
type MockStore struct {
	mock.Mock
}

func NewMockStore() *MockStore {
	return &MockStore{}
}

func (m *MockStore) GetAll() ([]models.Server, error) {
	args := m.Called()
	return args.Get(0).([]models.Server), args.Error(1)
}

func (m *MockStore) GetByID(id string) (*models.Server, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Server), args.Error(1)
}

func (m *MockStore) Create(server models.Server) error {
	args := m.Called(server)
	return args.Error(0)
}

func (m *MockStore) Update(server models.Server) error {
	args := m.Called(server)
	return args.Error(0)
}

func (m *MockStore) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockStore) Search(nameQuery string) ([]models.Server, error) {
	args := m.Called(nameQuery)
	return args.Get(0).([]models.Server), args.Error(1)
}

func (m *MockStore) Count() (total int, active int, err error) {
	args := m.Called()
	return args.Int(0), args.Int(1), args.Error(2)
}

func (m *MockStore) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Helper methods for setting up mock expectations

func (m *MockStore) ExpectGetAll(servers []models.Server, err error) *mock.Call {
	return m.On("GetAll").Return(servers, err)
}

func (m *MockStore) ExpectGetByID(id string, server *models.Server, err error) *mock.Call {
	return m.On("GetByID", id).Return(server, err)
}

func (m *MockStore) ExpectCreate(server models.Server, err error) *mock.Call {
	return m.On("Create", server).Return(err)
}

func (m *MockStore) ExpectUpdate(server models.Server, err error) *mock.Call {
	return m.On("Update", server).Return(err)
}

func (m *MockStore) ExpectDelete(id string, err error) *mock.Call {
	return m.On("Delete", id).Return(err)
}

func (m *MockStore) ExpectSearch(query string, servers []models.Server, err error) *mock.Call {
	return m.On("Search", query).Return(servers, err)
}

func (m *MockStore) ExpectCount(total, active int, err error) *mock.Call {
	return m.On("Count").Return(total, active, err)
}
