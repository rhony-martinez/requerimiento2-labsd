package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"pasarela-pagos/config"
)

// SimuladorFallos hace que las primeras CANTIDAD_FALLOS peticiones de pago de
// cada token de idempotencia "se pierdan":
//
//   - PETICION:  el mensaje de petición nunca llega a la lógica de negocio.
//     El pago NO se procesa.
//   - RESPUESTA: el pago SÍ se procesa, pero el mensaje de respuesta se pierde.
//
// En ambos casos el servidor retiene la conexión RETARDO_FALLO_MS milisegundos
// (más que el tiempo de espera del cliente) y luego la cierra sin responder.
type SimuladorFallos struct {
	modo     string
	cantidad int
	retardo  time.Duration

	mu       sync.Mutex
	intentos map[string]int
}

func NuevoSimuladorFallos(cfg config.Config) *SimuladorFallos {
	return &SimuladorFallos{
		modo:     cfg.ModoFallo,
		cantidad: cfg.CantidadFallos,
		retardo:  cfg.RetardoFallo,
		intentos: map[string]int{},
	}
}

func (s *SimuladorFallos) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if s.modo == config.ModoNinguno || s.cantidad == 0 {
			c.Next()
			return
		}

		token := leerToken(c)
		intento := s.registrarIntento(token)
		if intento > s.cantidad {
			log.Printf("[SIMULADOR] Token %s, petición #%d: sin fallo (ya se simularon %d)", token, intento, s.cantidad)
			c.Next()
			return
		}

		switch s.modo {
		case config.ModoPeticion:
			log.Printf("[SIMULADOR] Token %s, petición #%d: MENSAJE DE PETICIÓN PERDIDO (no se procesa el pago)", token, intento)
			time.Sleep(s.retardo)
			cortarConexion(c)
			c.Abort()

		case config.ModoRespuesta:
			original := c.Writer
			c.Writer = &escritorDescartable{ResponseWriter: original}
			c.Next() // el pago se procesa normalmente...
			c.Writer = original
			log.Printf("[SIMULADOR] Token %s, petición #%d: pago procesado, MENSAJE DE RESPUESTA PERDIDO", token, intento)
			time.Sleep(s.retardo)
			cortarConexion(c) // ...pero la respuesta nunca llega al cliente
		}
	}
}

func (s *SimuladorFallos) registrarIntento(token string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.intentos[token]++
	return s.intentos[token]
}

// leerToken extrae el token del cuerpo y lo restaura para que el controlador pueda leerlo.
func leerToken(c *gin.Context) string {
	cuerpo, _ := io.ReadAll(c.Request.Body)
	c.Request.Body = io.NopCloser(bytes.NewReader(cuerpo))
	var sol struct {
		Token string `json:"token_idempotencia"`
	}
	_ = json.Unmarshal(cuerpo, &sol)
	return sol.Token
}

// cortarConexion cierra la conexión TCP sin enviar ninguna respuesta HTTP.
func cortarConexion(c *gin.Context) {
	conexion, _, err := c.Writer.Hijack()
	if err != nil {
		log.Printf("[SIMULADOR] No se pudo cerrar la conexión: %v", err)
		return
	}
	_ = conexion.Close()
}

// escritorDescartable descarta todo lo que el controlador intente escribir.
type escritorDescartable struct {
	gin.ResponseWriter
}

func (e *escritorDescartable) WriteHeader(int)                   {}
func (e *escritorDescartable) WriteHeaderNow()                   {}
func (e *escritorDescartable) Write(b []byte) (int, error)       { return len(b), nil }
func (e *escritorDescartable) WriteString(s string) (int, error) { return len(s), nil }
func (e *escritorDescartable) Flush()                            {}
