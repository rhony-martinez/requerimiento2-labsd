package services

import (
	"time"

	"pasarela-pagos/data/entities"
	"pasarela-pagos/services/dto"
)

// Funciones de mapeo entre entidades (capa de datos) y DTO (capa de controladores).

const formatoFecha = time.RFC3339

func aSuscripcionDTO(s entities.Suscripcion) dto.SuscripcionDTO {
	return dto.SuscripcionDTO{
		ID:          s.ID,
		Nombre:      s.Nombre,
		Descripcion: s.Descripcion,
		Precio:      s.Precio,
		Moneda:      s.Moneda,
		Cuentas:     s.Cuentas,
	}
}

func aRespuestaTokenDTO(t entities.TokenIdempotencia) dto.RespuestaTokenDTO {
	return dto.RespuestaTokenDTO{
		TokenIdempotencia: t.Token,
		Cedula:            t.Cedula,
		SuscripcionID:     t.SuscripcionID,
		Monto:             t.Monto,
		FechaCreacion:     t.FechaCreacion.Format(formatoFecha),
	}
}

func aTokenDTO(t entities.TokenIdempotencia) dto.TokenDTO {
	return dto.TokenDTO{
		TokenIdempotencia: t.Token,
		SuscripcionID:     t.SuscripcionID,
		Monto:             t.Monto,
		FechaCreacion:     t.FechaCreacion.Format(formatoFecha),
		Utilizado:         t.Utilizado,
	}
}

func aTransaccionDTO(t entities.Transaccion) dto.TransaccionDTO {
	return dto.TransaccionDTO{
		IDTransaccion:     t.ID,
		TokenIdempotencia: t.Token,
		SuscripcionID:     t.SuscripcionID,
		Monto:             t.Monto,
		Fecha:             t.Fecha.Format(formatoFecha),
	}
}
