package services

import (
	"pasarela-pagos/data/repositories"
	"pasarela-pagos/services/dto"
)

type SuscripcionService struct {
	repo *repositories.SuscripcionRepository
}

func NuevoSuscripcionService(repo *repositories.SuscripcionRepository) *SuscripcionService {
	return &SuscripcionService{repo: repo}
}

func (s *SuscripcionService) Listar() []dto.SuscripcionDTO {
	resultado := []dto.SuscripcionDTO{}
	for _, sus := range s.repo.ListarTodas() {
		resultado = append(resultado, aSuscripcionDTO(sus))
	}
	return resultado
}
