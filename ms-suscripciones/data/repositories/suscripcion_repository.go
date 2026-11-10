package repositories

import "pasarela-pagos/data/entities"

// SuscripcionRepository mantiene en memoria el catálogo de planes de Spotify.
type SuscripcionRepository struct {
	suscripciones []entities.Suscripcion
}

func NuevoSuscripcionRepository() *SuscripcionRepository {
	return &SuscripcionRepository{
		suscripciones: []entities.Suscripcion{
			{ID: "individual", Nombre: "Premium Individual", Descripcion: "Música sin anuncios, descargas y saltos ilimitados para 1 cuenta.", Precio: 16900, Moneda: "COP", Cuentas: 1},
			{ID: "duo", Nombre: "Premium Duo", Descripcion: "2 cuentas Premium para personas que viven juntas y Duo Mix.", Precio: 21900, Moneda: "COP", Cuentas: 2},
			{ID: "familiar", Nombre: "Premium Familiar", Descripcion: "Hasta 6 cuentas Premium, control parental y Family Mix.", Precio: 26900, Moneda: "COP", Cuentas: 6},
			{ID: "estudiantes", Nombre: "Premium Estudiantes", Descripcion: "Precio especial para estudiantes universitarios verificados.", Precio: 8900, Moneda: "COP", Cuentas: 1},
		},
	}
}

func (r *SuscripcionRepository) ListarTodas() []entities.Suscripcion {
	copia := make([]entities.Suscripcion, len(r.suscripciones))
	copy(copia, r.suscripciones)
	return copia
}

func (r *SuscripcionRepository) BuscarPorID(id string) (entities.Suscripcion, bool) {
	for _, s := range r.suscripciones {
		if s.ID == id {
			return s, true
		}
	}
	return entities.Suscripcion{}, false
}
