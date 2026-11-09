package repositories

import (
	"sync"

	"ms-estadisticas/entities"
)

type ReproduccionRepository struct {
	mu    sync.Mutex
	datos []entities.Reproduccion
}

func NewReproduccionRepository() *ReproduccionRepository {
	return &ReproduccionRepository{datos: make([]entities.Reproduccion, 0)}
}

func (r *ReproduccionRepository) Guardar(rep entities.Reproduccion) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.datos = append(r.datos, rep)
}

func (r *ReproduccionRepository) ListarPorUsuario(usuarioID int) []entities.Reproduccion {
	r.mu.Lock()
	defer r.mu.Unlock()
	resultado := make([]entities.Reproduccion, 0)
	for _, rep := range r.datos {
		if rep.UsuarioID == usuarioID {
			resultado = append(resultado, rep)
		}
	}
	return resultado
}
