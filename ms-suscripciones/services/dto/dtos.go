package dto

// ---------- Suscripciones ----------

type SuscripcionDTO struct {
	ID          string  `json:"id"`
	Nombre      string  `json:"nombre"`
	Descripcion string  `json:"descripcion"`
	Precio      float64 `json:"precio"`
	Moneda      string  `json:"moneda"`
	Cuentas     int     `json:"cuentas"`
}

// ---------- Servicio 1: generar token de idempotencia ----------

type SolicitudTokenDTO struct {
	Cedula        string  `json:"cedula"`
	SuscripcionID string  `json:"suscripcion_id"`
	Monto         float64 `json:"monto"`
	NombreTitular string  `json:"nombre_titular"`
	NumeroTarjeta string  `json:"numero_tarjeta"`
}

type RespuestaTokenDTO struct {
	TokenIdempotencia string  `json:"token_idempotencia"`
	Cedula            string  `json:"cedula"`
	SuscripcionID     string  `json:"suscripcion_id"`
	Monto             float64 `json:"monto"`
	FechaCreacion     string  `json:"fecha_creacion"`
}

// ---------- Servicio 2: realizar pago ----------

type SolicitudPagoDTO struct {
	TokenIdempotencia string  `json:"token_idempotencia"`
	Monto             float64 `json:"monto"`
}

type RespuestaPagoDTO struct {
	TokenIdempotencia      string   `json:"token_idempotencia"`
	MontoTransaccion       float64  `json:"monto_transaccion"`
	TransaccionRealizada   bool     `json:"transaccion_realizada"`
	PreviamenteRealizada   bool     `json:"previamente_realizada"`
	MontoTransaccionPrevia *float64 `json:"monto_transaccion_previa,omitempty"`
	IDTransaccion          string   `json:"id_transaccion,omitempty"`
	FechaTransaccion       string   `json:"fecha_transaccion,omitempty"`
	Mensaje                string   `json:"mensaje"`
}

// ---------- Servicio 3: consultar transacciones ----------

type TransaccionDTO struct {
	IDTransaccion     string  `json:"id_transaccion"`
	TokenIdempotencia string  `json:"token_idempotencia"`
	SuscripcionID     string  `json:"suscripcion_id"`
	Monto             float64 `json:"monto"`
	Fecha             string  `json:"fecha"`
}

type TokenDTO struct {
	TokenIdempotencia string  `json:"token_idempotencia"`
	SuscripcionID     string  `json:"suscripcion_id"`
	Monto             float64 `json:"monto"`
	FechaCreacion     string  `json:"fecha_creacion"`
	Utilizado         bool    `json:"utilizado"`
}

type RespuestaConsultaDTO struct {
	Cedula             string           `json:"cedula"`
	TotalTransacciones int              `json:"total_transacciones"`
	Transacciones      []TransaccionDTO `json:"transacciones"`
	TokensUtilizados   []TokenDTO       `json:"tokens_utilizados"`
	TokensSinUtilizar  []TokenDTO       `json:"tokens_sin_utilizar"`
}

// ---------- Errores ----------

type ErrorDTO struct {
	Error string `json:"error"`
}
