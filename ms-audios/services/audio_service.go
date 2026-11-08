package services

import (
	"ms-audios/dto"
	"ms-audios/entities"
	"ms-audios/repositories"
)

type AudioService struct {
	repo *repositories.AudioRepository
}

func NewAudioService(repo *repositories.AudioRepository) *AudioService {
	return &AudioService{repo: repo}
}

func (s *AudioService) ListarTipos() []dto.TipoAudioDTO {
	tipos := s.repo.ListarTipos()
	resultado := make([]dto.TipoAudioDTO, 0, len(tipos))
	for _, t := range tipos {
		resultado = append(resultado, dto.TipoAudioDTO{ID: t.ID, Nombre: t.Nombre})
	}
	return resultado
}

func (s *AudioService) ListarAudiosPorTipo(tipoID int) ([]dto.AudioResumenDTO, error) {
	audios, err := s.repo.ListarAudiosPorTipo(tipoID)
	if err != nil {
		return nil, err
	}
	resultado := make([]dto.AudioResumenDTO, 0, len(audios))
	for _, a := range audios {
		if a.TipoID == tipoID {
			resultado = append(resultado, dto.AudioResumenDTO{ID: a.ID, Titulo: a.Titulo})
		}
	}
	return resultado, nil
}

func (s *AudioService) ObtenerDetalle(id int) (dto.AudioDetalleDTO, error) {
	audio, err := s.repo.ObtenerAudio(id)
	if err != nil {
		return dto.AudioDetalleDTO{}, err
	}
	tipo, err := s.repo.ObtenerTipo(audio.TipoID)
	if err != nil {
		return dto.AudioDetalleDTO{}, err
	}
	return toDetalleDTO(audio, tipo.Nombre), nil
}

func toDetalleDTO(a entities.Audio, nombreTipo string) dto.AudioDetalleDTO {
	return dto.AudioDetalleDTO{
		ID: a.ID, Tipo: nombreTipo, Titulo: a.Titulo, Archivo: a.Archivo, Formato: a.Formato, DuracionSeg: a.DuracionSeg, Metadatos: a.Metadatos,
	}
}
