package repositories

import (
	"sync"

	"pasarela-pagos/data/entities"
)

// TransaccionRepository es el historial de transacciones realizadas exitosamente.
type TransaccionRepository struct {
	mu            sync.RWMutex
	ruta          string
	transacciones []entities.Transaccion
}

func NuevoTransaccionRepository(ruta string) (*TransaccionRepository, error) {
	r := &TransaccionRepository{ruta: ruta, transacciones: []entities.Transaccion{}}
	if err := cargarJSON(ruta, &r.transacciones); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *TransaccionRepository) Guardar(t entities.Transaccion) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.transacciones = append(r.transacciones, t)
	return guardarJSON(r.ruta, r.transacciones)
}

// BuscarPorToken indica si ya existe una transacción realizada con el token dado.
func (r *TransaccionRepository) BuscarPorToken(token string) (entities.Transaccion, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, t := range r.transacciones {
		if t.Token == token {
			return t, true
		}
	}
	return entities.Transaccion{}, false
}

func (r *TransaccionRepository) ListarPorCedula(cedula string) []entities.Transaccion {
	r.mu.RLock()
	defer r.mu.RUnlock()
	resultado := []entities.Transaccion{}
	for _, t := range r.transacciones {
		if t.Cedula == cedula {
			resultado = append(resultado, t)
		}
	}
	return resultado
}
