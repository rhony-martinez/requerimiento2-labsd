package repositories

import (
	"errors"
	"log"
	"os"
	"sync"

	"ms-audios/entities"
)

var (
	ErrTipoNoEncontrado  = errors.New("el tipo de audio no existe")
	ErrAudioNoEncontrado = errors.New("el audio no existe")
)

type AudioRepository struct {
	mu     sync.RWMutex
	ruta   string
	tipos  []entities.TipoAudio
	audios []entities.Audio
}

func NewAudioRepository(ruta string) *AudioRepository {
	repo := &AudioRepository{ruta: ruta, tipos: []entities.TipoAudio{}, audios: []entities.Audio{}}

	contenido, err := leerJSON(ruta)
	switch {
	case errors.Is(err, os.ErrNotExist):
		// Primer arranque: sembramos y persistimos
		repo.cargarDatosIniciales()
		if errGuardar := repo.guardar(); errGuardar != nil {
			log.Printf("no se pudo crear %s: %v", ruta, errGuardar)
		}
	case err != nil:
		// Existe pero está dañado, evitamos perder datos
		log.Fatalf("archivo de datos inválido (%s): %v", ruta, err)
	default:
		repo.tipos = contenido.Tipos
		repo.audios = contenido.Audios
	}
	return repo
}

// guardar persiste el estado actual. Debe invocarse con lock de escritura o en construcción
func (r *AudioRepository) guardar() error {
	return escribirJSON(r.ruta, contenidoJSON{Tipos: r.tipos, Audios: r.audios})
}

func (r *AudioRepository) ListarTipos() []entities.TipoAudio {
	r.mu.RLock()
	defer r.mu.RUnlock()
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
