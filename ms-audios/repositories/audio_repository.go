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
	ErrTituloDuplicado   = errors.New("ya existe un audio con ese título en este tipo")
)

type AudioRepository struct {
	mu          sync.RWMutex
	ruta        string
	tipos       []entities.TipoAudio
	audios      []entities.Audio
	siguienteID int
}

func NewAudioRepository(ruta string) *AudioRepository {
	repo := &AudioRepository{ruta: ruta, tipos: []entities.TipoAudio{}, audios: []entities.Audio{}}

	contenido, err := leerJSON(ruta)
	switch {
	case errors.Is(err, os.ErrNotExist):
		// Primer arranque: sembramos y persistimos
		repo.cargarDatosIniciales()
		repo.corregirSiguienteID()
		if errGuardar := repo.guardar(); errGuardar != nil {
			log.Printf("no se pudo crear %s: %v", ruta, errGuardar)
		}
	case err != nil:
		// Existe pero está dañado, evitamos perder datos
		log.Fatalf("archivo de datos inválido (%s): %v", ruta, err)
	default:
		repo.tipos = contenido.Tipos
		repo.audios = contenido.Audios
		repo.siguienteID = contenido.SiguienteID
		repo.corregirSiguienteID()
	}
	return repo
}

// guardar persiste el estado actual, el contador
func (r *AudioRepository) guardar() error {
	return escribirJSON(r.ruta, contenidoJSON{Tipos: r.tipos, Audios: r.audios, SiguienteID: r.siguienteID})
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

// Agregar asigna el id, persiste y, si falla el guardado, deja la memoria como estaba.
func (r *AudioRepository) Agregar(a entities.Audio) (entities.Audio, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, existente := range r.audios {
		if existente.TipoID == a.TipoID && existente.Titulo == a.Titulo {
			return entities.Audio{}, ErrTituloDuplicado
		}
	}
	a.ID = r.siguienteID
	r.siguienteID++
	r.audios = append(r.audios, a)
	if err := r.guardar(); err != nil {
		r.audios = r.audios[:len(r.audios)-1] // deshacer
		r.siguienteID--
		return entities.Audio{}, err
	}
	return a, nil
}

// Eliminar quita el audio del catálogo. Devuelve el audio borrado y si otro
// audio sigue usando el mismo archivo en disco.
func (r *AudioRepository) Eliminar(id int) (entities.Audio, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	pos := -1
	for i, a := range r.audios {
		if a.ID == id {
			pos = i
			break
		}
	}
	if pos == -1 {
		return entities.Audio{}, false, ErrAudioNoEncontrado
	}
	borrado := r.audios[pos]

	// Slice nuevo en vez de append sobre el original: así, si guardar falla,
	// el anterior sigue intacto y deshacer es una simple asignación.
	nuevos := make([]entities.Audio, 0, len(r.audios)-1)
	nuevos = append(nuevos, r.audios[:pos]...)
	nuevos = append(nuevos, r.audios[pos+1:]...)
	anteriores := r.audios
	r.audios = nuevos
	if err := r.guardar(); err != nil {
		r.audios = anteriores
		return entities.Audio{}, false, err
	}

	compartido := false
	for _, a := range r.audios {
		if a.Archivo == borrado.Archivo && a.Formato == borrado.Formato {
			compartido = true
			break
		}
	}
	return borrado, compartido, nil
}

func (r *AudioRepository) corregirSiguienteID() {
	for _, a := range r.audios {
		if a.ID >= r.siguienteID {
			r.siguienteID = a.ID + 1
		}
	}
	if r.siguienteID < 1 {
		r.siguienteID = 1
	}
}
