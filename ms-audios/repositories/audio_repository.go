package repositories

import (
	"errors"
	"sync"

	"ms-audios/entities"
)

var (
	ErrTipoNoEncontrado  = errors.New("el tipo de audio no existe")
	ErrAudioNoEncontrado = errors.New("el audio no existe")
)

type AudioRepository struct {
	mu     sync.RWMutex
	tipos  []entities.TipoAudio
	audios []entities.Audio
}

func NewAudioRepository() *AudioRepository {
	repo := &AudioRepository{}
	repo.cargarDatosIniciales()
	return repo
}

func (r *AudioRepository) ListarTipos() []entities.TipoAudio {
	r.mu.RLock()
	defer r.mu.Unlock()
	copia := make([]entities.TipoAudio, len(r.tipos))
	copy(copia, r.tipos)
	return copia
}

func (r *AudioRepository) ObtenerTipo(id int) (entities.TipoAudio, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, t := range r.tipos {
		if t.ID == id {
			return t, nil
		}
	}
	return entities.TipoAudio{}, ErrTipoNoEncontrado
}

func (r *AudioRepository) ListarAudiosPorTipo(tipoID int) ([]entities.Audio, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	existe := false
	for _, t := range r.tipos {
		if t.ID == tipoID {
			existe = true
			break
		}
	}

	if !existe {
		return nil, ErrTipoNoEncontrado
	}

	resultado := make([]entities.Audio, 0)
	for _, a := range r.audios {
		if a.TipoID == tipoID {
			resultado = append(resultado, a)
		}
	}
	return resultado, nil
}

func (r *AudioRepository) ObtenerAudio(id int) (entities.Audio, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, a := range r.audios {
		if a.ID == id {
			return a, nil
		}
	}
	return entities.Audio{}, ErrAudioNoEncontrado
}
