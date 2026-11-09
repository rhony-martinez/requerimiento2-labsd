package services

import (
	"errors"
	"sort"

	"ms-estadisticas/dto"
	"ms-estadisticas/entities"
	"ms-estadisticas/repositories"
)

var ErrMensajeInvalido = errors.New("mensaje de reproducción inválido")

type EstadisticasService struct {
	repo *repositories.ReproduccionRepository
}

func NewEstadisticasService(repo *repositories.ReproduccionRepository) *EstadisticasService {
	return &EstadisticasService{repo: repo}
}

// RegistarReproduccion la invoca el consumidor de la cola
func (s *EstadisticasService) RegistrarReproduccion(m dto.ReproduccionMensajeDTO) error {
	if m.UsuarioID <= 0 || m.AudioID <= 0 {
		return ErrMensajeInvalido
	}
	return s.repo.Guardar(entities.Reproduccion{UsuarioID: m.UsuarioID, AudioID: m.AudioID, Fecha: m.Fecha})
}

// HistorialDelUsuario agrupa los eventos por audio y cuenta cuántas veces se reprodujo cada uno
func (s *EstadisticasService) HistorialDelUsuario(usuarioID int) []dto.HistorialItemDTO {
	acumulado := make(map[int]*dto.HistorialItemDTO)
	for _, rep := range s.repo.ListarPorUsuario(usuarioID) {
		item, existe := acumulado[rep.AudioID]
		if !existe {
			item = &dto.HistorialItemDTO{AudioID: rep.AudioID}
			acumulado[rep.AudioID] = item
		}
		item.Reproducciones++
		if rep.Fecha.After(item.UltimaFecha) {
			item.UltimaFecha = rep.Fecha
		}
	}

	resultado := make([]dto.HistorialItemDTO, 0, len(acumulado)) // nunca nil: el JSON sale []
	for _, item := range acumulado {
		resultado = append(resultado, *item)
	}

	// El mapa no tiene orden: ordenamos por lo más reciente primero
	sort.Slice(resultado, func(i, j int) bool {
		return resultado[i].UltimaFecha.After(resultado[j].UltimaFecha)
	})
	return resultado
}
