package service

import "github.com/janobono/auth-service/internal/repository"

type AttributeService interface {
}

type attributeService struct {
	attributeRepository repository.AttributeRepository
}

var _ AttributeService = (*attributeService)(nil)

func NewAttributeService(attributeRepository repository.AttributeRepository) AttributeService {
	return &attributeService{attributeRepository}
}
