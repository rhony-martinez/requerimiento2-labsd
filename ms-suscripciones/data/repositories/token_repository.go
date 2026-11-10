package repositories

import (
	"sort"
	"sync"

	"pasarela-pagos/data/entities"
)

// TokenRepository almacena los tokens de idempotencia generados.
type TokenRepository struct {
	mu     sync.RWMutex
	ruta   string
	tokens map[string]entities.TokenIdempotencia
}

func NuevoTokenRepository(ruta string) (*TokenRepository, error) {
	r := &TokenRepository{ruta: ruta, tokens: map[string]entities.TokenIdempotencia{}}
	if err := cargarJSON(ruta, &r.tokens); err != nil {
		return nil, err
	}
	return r, nil
}

// Guardar inserta o actualiza un token y persiste el almacén.
func (r *TokenRepository) Guardar(t entities.TokenIdempotencia) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tokens[t.Token] = t
	return guardarJSON(r.ruta, r.tokens)
}

func (r *TokenRepository) BuscarPorToken(token string) (entities.TokenIdempotencia, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tokens[token]
	return t, ok
}

func (r *TokenRepository) ListarPorCedula(cedula string) []entities.TokenIdempotencia {
	r.mu.RLock()
	defer r.mu.RUnlock()
	resultado := []entities.TokenIdempotencia{}
	for _, t := range r.tokens {
		if t.Cedula == cedula {
			resultado = append(resultado, t)
		}
	}
	sort.Slice(resultado, func(i, j int) bool {
		return resultado[i].FechaCreacion.Before(resultado[j].FechaCreacion)
	})
	return resultado
}
