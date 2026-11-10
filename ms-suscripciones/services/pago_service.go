package services

import (
	"crypto/rand"
	"fmt"
	"log"
	"math"
	"regexp"
	"strings"
	"sync"
	"time"

	"pasarela-pagos/data/entities"
	"pasarela-pagos/data/repositories"
	"pasarela-pagos/services/dto"
)

var (
	regexCedula  = regexp.MustCompile(`^[0-9]{6,10}$`)
	regexTarjeta = regexp.MustCompile(`^[0-9]{13,19}$`)
)

// PagoService es la fachada de la pasarela: valida, aplica la regla de
// idempotencia y mapea entre entidades y DTO.
type PagoService struct {
	// mu hace atómico el "¿ya se usó el token? -> registrar transacción",
	// evitando que dos peticiones con el mismo token paguen dos veces.
	mu              sync.Mutex
	tokenRepo       *repositories.TokenRepository
	transaccionRepo *repositories.TransaccionRepository
	suscripcionRepo *repositories.SuscripcionRepository
}

func NuevoPagoService(
	tokenRepo *repositories.TokenRepository,
	transaccionRepo *repositories.TransaccionRepository,
	suscripcionRepo *repositories.SuscripcionRepository,
) *PagoService {
	return &PagoService{tokenRepo: tokenRepo, transaccionRepo: transaccionRepo, suscripcionRepo: suscripcionRepo}
}

// GenerarToken (servicio REST 1) valida los datos del pago y emite un token de idempotencia.
func (s *PagoService) GenerarToken(sol dto.SolicitudTokenDTO) (dto.RespuestaTokenDTO, error) {
	sol.Cedula = strings.TrimSpace(sol.Cedula)
	sol.NombreTitular = strings.TrimSpace(sol.NombreTitular)
	tarjeta := strings.ReplaceAll(strings.TrimSpace(sol.NumeroTarjeta), " ", "")

	if !regexCedula.MatchString(sol.Cedula) {
		return dto.RespuestaTokenDTO{}, errorValidacion("la cédula debe tener entre 6 y 10 dígitos")
	}
	if sol.NombreTitular == "" {
		return dto.RespuestaTokenDTO{}, errorValidacion("el nombre del titular es obligatorio")
	}
	if !regexTarjeta.MatchString(tarjeta) {
		return dto.RespuestaTokenDTO{}, errorValidacion("el número de tarjeta debe tener entre 13 y 19 dígitos")
	}
	suscripcion, ok := s.suscripcionRepo.BuscarPorID(sol.SuscripcionID)
	if !ok {
		return dto.RespuestaTokenDTO{}, errorNoEncontrado(fmt.Sprintf("la suscripción '%s' no existe", sol.SuscripcionID))
	}
	if !mismoMonto(sol.Monto, suscripcion.Precio) {
		return dto.RespuestaTokenDTO{}, errorNegocio(fmt.Sprintf("el monto %.0f no corresponde al precio de la suscripción (%.0f)", sol.Monto, suscripcion.Precio))
	}

	token := entities.TokenIdempotencia{
		Token:                 nuevoUUID(),
		Cedula:                sol.Cedula,
		SuscripcionID:         suscripcion.ID,
		Monto:                 suscripcion.Precio,
		NombreTitular:         sol.NombreTitular,
		UltimosDigitosTarjeta: tarjeta[len(tarjeta)-4:],
		FechaCreacion:         time.Now(),
	}
	if err := s.tokenRepo.Guardar(token); err != nil {
		return dto.RespuestaTokenDTO{}, errorInterno("no fue posible almacenar el token: " + err.Error())
	}
	log.Printf("[TOKEN] Generado %s para cédula %s (%s, $%.0f)", token.Token, token.Cedula, token.SuscripcionID, token.Monto)
	return aRespuestaTokenDTO(token), nil
}

// RealizarPago (servicio REST 2) ejecuta el pago una única vez por token de idempotencia.
func (s *PagoService) RealizarPago(sol dto.SolicitudPagoDTO) (dto.RespuestaPagoDTO, error) {
	sol.TokenIdempotencia = strings.TrimSpace(sol.TokenIdempotencia)
	if sol.TokenIdempotencia == "" {
		return dto.RespuestaPagoDTO{}, errorValidacion("el token de idempotencia es obligatorio")
	}
	if sol.Monto <= 0 {
		return dto.RespuestaPagoDTO{}, errorValidacion("el monto debe ser mayor que cero")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	token, ok := s.tokenRepo.BuscarPorToken(sol.TokenIdempotencia)
	if !ok {
		return dto.RespuestaPagoDTO{}, errorNoEncontrado("el token de idempotencia no existe")
	}

	// Regla de idempotencia: antes de pagar se pregunta si el token ya fue utilizado.
	if previa, existe := s.transaccionRepo.BuscarPorToken(token.Token); existe {
		log.Printf("[PAGO] Token %s YA UTILIZADO (transacción %s por $%.0f). No se re-ejecuta el pago.", token.Token, previa.ID, previa.Monto)
		montoPrevio := previa.Monto
		return dto.RespuestaPagoDTO{
			TokenIdempotencia:      token.Token,
			MontoTransaccion:       sol.Monto,
			TransaccionRealizada:   false,
			PreviamenteRealizada:   true,
			MontoTransaccionPrevia: &montoPrevio,
			IDTransaccion:          previa.ID,
			FechaTransaccion:       previa.Fecha.Format(formatoFecha),
			Mensaje:                "El pago no se ejecutó de nuevo: ya existía una transacción realizada con este token de idempotencia",
		}, nil
	}

	if !mismoMonto(sol.Monto, token.Monto) {
		return dto.RespuestaPagoDTO{}, errorNegocio(fmt.Sprintf("el monto %.0f no coincide con el monto registrado para el token (%.0f)", sol.Monto, token.Monto))
	}

	transaccion := entities.Transaccion{
		ID:            nuevoUUID(),
		Token:         token.Token,
		Cedula:        token.Cedula,
		SuscripcionID: token.SuscripcionID,
		Monto:         sol.Monto,
		Fecha:         time.Now(),
	}
	if err := s.transaccionRepo.Guardar(transaccion); err != nil {
		return dto.RespuestaPagoDTO{}, errorInterno("no fue posible registrar la transacción: " + err.Error())
	}
	token.Utilizado = true
	if err := s.tokenRepo.Guardar(token); err != nil {
		return dto.RespuestaPagoDTO{}, errorInterno("no fue posible actualizar el token: " + err.Error())
	}
	log.Printf("[PAGO] Pago ALMACENADO: transacción %s guardada y token %s marcado como utilizado", transaccion.ID, token.Token)
	log.Printf("[PAGO] Transacción %s REALIZADA: cédula %s, $%.0f, token %s", transaccion.ID, transaccion.Cedula, transaccion.Monto, token.Token)
	return dto.RespuestaPagoDTO{
		TokenIdempotencia:    token.Token,
		MontoTransaccion:     transaccion.Monto,
		TransaccionRealizada: true,
		PreviamenteRealizada: false,
		IDTransaccion:        transaccion.ID,
		FechaTransaccion:     transaccion.Fecha.Format(formatoFecha),
		Mensaje:              "Pago realizado exitosamente",
	}, nil
}

// ConsultarTransacciones (servicio REST 3) devuelve el historial de pagos y tokens de una cédula.
func (s *PagoService) ConsultarTransacciones(cedula string) (dto.RespuestaConsultaDTO, error) {
	cedula = strings.TrimSpace(cedula)
	if !regexCedula.MatchString(cedula) {
		return dto.RespuestaConsultaDTO{}, errorValidacion("la cédula debe tener entre 6 y 10 dígitos")
	}

	respuesta := dto.RespuestaConsultaDTO{
		Cedula:            cedula,
		Transacciones:     []dto.TransaccionDTO{},
		TokensUtilizados:  []dto.TokenDTO{},
		TokensSinUtilizar: []dto.TokenDTO{},
	}
	for _, t := range s.transaccionRepo.ListarPorCedula(cedula) {
		respuesta.Transacciones = append(respuesta.Transacciones, aTransaccionDTO(t))
	}
	for _, t := range s.tokenRepo.ListarPorCedula(cedula) {
		if t.Utilizado {
			respuesta.TokensUtilizados = append(respuesta.TokensUtilizados, aTokenDTO(t))
		} else {
			respuesta.TokensSinUtilizar = append(respuesta.TokensSinUtilizar, aTokenDTO(t))
		}
	}
	respuesta.TotalTransacciones = len(respuesta.Transacciones)
	return respuesta, nil
}

func mismoMonto(a, b float64) bool {
	return math.Abs(a-b) < 0.005
}

// nuevoUUID genera un identificador UUID v4.
func nuevoUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
