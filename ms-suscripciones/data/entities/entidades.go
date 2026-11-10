package entities

import "time"

// Suscripcion representa un plan de la plataforma Spotify.
type Suscripcion struct {
	ID          string  `json:"id"`
	Nombre      string  `json:"nombre"`
	Descripcion string  `json:"descripcion"`
	Precio      float64 `json:"precio"`
	Moneda      string  `json:"moneda"`
	Cuentas     int     `json:"cuentas"`
}

// TokenIdempotencia se genera antes de pagar y asocia al cliente con los datos del pago.
type TokenIdempotencia struct {
	Token                 string    `json:"token"`
	Cedula                string    `json:"cedula"`
	SuscripcionID         string    `json:"suscripcion_id"`
	Monto                 float64   `json:"monto"`
	NombreTitular         string    `json:"nombre_titular"`
	UltimosDigitosTarjeta string    `json:"ultimos_digitos_tarjeta"`
	FechaCreacion         time.Time `json:"fecha_creacion"`
	Utilizado             bool      `json:"utilizado"`
}

// Transaccion es un pago realizado exitosamente (historial de transacciones).
type Transaccion struct {
	ID            string    `json:"id"`
	Token         string    `json:"token"`
	Cedula        string    `json:"cedula"`
	SuscripcionID string    `json:"suscripcion_id"`
	Monto         float64   `json:"monto"`
	Fecha         time.Time `json:"fecha"`
}
