package service

import (
	"github.com/digitalpapyrus/backend/internal/model"
	"github.com/digitalpapyrus/backend/internal/repository"
)

type CoreServiceService struct {
	repo *repository.CoreServiceRepository
}

func NewCoreServiceService(repo *repository.CoreServiceRepository) *CoreServiceService {
	return &CoreServiceService{repo: repo}
}

func (s *CoreServiceService) FindAll() ([]*model.CoreService, error) {
	return s.repo.FindAll()
}

func (s *CoreServiceService) FindByID(id string) (*model.CoreService, error) {
	return s.repo.FindByID(id)
}

func (s *CoreServiceService) Create(cs *model.CoreService) error {
	return s.repo.Create(cs)
}

func (s *CoreServiceService) Update(cs *model.CoreService) error {
	return s.repo.Update(cs)
}

func (s *CoreServiceService) Delete(id string) error {
	return s.repo.Delete(id)
}
