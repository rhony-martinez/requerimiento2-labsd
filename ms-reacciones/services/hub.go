package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"golang.org/x/time/rate"

	"ms-reacciones/dto"
)

var (
	ErrReaccionInvalida     = errors.New("reacción no permitida")
	ErrSinSala              = errors.New("primero debes unirte a un audio (reproduciéndolo)")
	ErrAudioInvalido        = errors.New("audio_id inválido")
	ErrDemasiadasReacciones = errors.New("demasiadas reacciones seguidas; espera un momento")
)

// Lista blanca: nunca reenviamos texto arbitrario que escriba un cliente
var reaccionesPermitidas = map[string]bool{
	"like": true, "heart": true, "fun": true, "clap": true, "fire": true,
}

// Cliente representa una conexión. No conoce WebSocket, solo un buzón.
type Cliente struct {
	UsuarioID int
	Envio     chan []byte   // mensajes pendientes de entregar
	Cerrado   chan struct{} // se cierra al desconectarse
	audioID   int           // 0 = en ninguna sala; protegido por Hub.mu
	limitador *rate.Limiter // fichas propias de esta conexión
}

func NuevoCliente(usuarioID int) *Cliente {
	return &Cliente{
		UsuarioID: usuarioID,
		Envio:     make(chan []byte, 16),
		Cerrado:   make(chan struct{}),
		limitador: rate.NewLimiter(rate.Limit(5), 5), // 5 fichas/s
	}
}

// Enviar encola sin bloquear. Si el buzón está lleno (cliente lento) descarta:
// una reacción perdida no vale que un cliente lento frene a los demás.
// Envio nunca se cierra, así que enviar tras desconectarse no provoca panic.
func (c *Cliente) Enviar(msg []byte) bool {
	select {
	case c.Envio <- msg:
		return true
	default:
		return false
	}
}

type Hub struct {
	mu    sync.RWMutex
	salas map[int]map[*Cliente]struct{} // audioID -> clientes
}

func NuevoHub() *Hub {
	return &Hub{salas: make(map[int]map[*Cliente]struct{})}
}

// Unir mete al cliente en la sala de un audio (saliendo antes de la anterior).
func (h *Hub) Unir(c *Cliente, audioID int) error {
	if audioID <= 0 {
		return ErrAudioInvalido
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if c.audioID == audioID {
		return nil // ya estaba en esa sala
	}

	anterior := h.quitarSinLock(c) // sala de la que venía (0 si ninguna)

	if h.salas[audioID] == nil {
		h.salas[audioID] = make(map[*Cliente]struct{})
	}
	h.salas[audioID][c] = struct{}{}
	c.audioID = audioID

	h.notificarOyentesSinLock(anterior) // los que se quedaron atrás
	h.notificarOyentesSinLock(audioID)  // los de la sala nueva, incluido c
	return nil
}

func (h *Hub) Salir(c *Cliente) {
	h.mu.Lock()
	defer h.mu.Unlock()
	anterior := h.quitarSinLock(c)
	h.notificarOyentesSinLock(anterior)
}

// quitarSinLock exige tener h.mu tomado. Es idempotente.
func (h *Hub) quitarSinLock(c *Cliente) int {
	anterior := c.audioID
	if anterior == 0 {
		return 0
	}
	if sala := h.salas[anterior]; sala != nil {
		delete(sala, c)
		if len(sala) == 0 {
			delete(h.salas, c.audioID) // no acumulamos salas vacías
		}
	}
	c.audioID = 0
	return anterior
}

func (h *Hub) notificarOyentesSinLock(audioID int) {

	sala := h.salas[audioID]
	if audioID == 0 || len(sala) == 0 {
		return
	}

	cantidad := len(sala)

	cuerpo, _ := json.Marshal(dto.MensajeSaliente{
		Tipo: "oyentes", AudioID: audioID, Cantidad: cantidad,
	})
	for c := range sala {
		c.Enviar(cuerpo)
	}

	fmt.Printf("[WS] sala del audio %d: %d oyente(s)\n", audioID, cantidad)
}

// Reaccionar reenvía la reacción a los demás clientes de la sala del remitente
// Devuelve a cuántos se la entregó y el audio al que pertenece
func (h *Hub) Reacciones(origen *Cliente, reaccion string) (entregadas int, audioID int, err error) {
	if !origen.limitador.Allow() {
		return 0, 0, ErrDemasiadasReacciones
	}
	if !reaccionesPermitidas[reaccion] {
		return 0, 0, ErrReaccionInvalida
	}
	h.mu.RLock()
	defer h.mu.RUnlock()

	audioID = origen.audioID
	if audioID == 0 {
		return 0, 0, ErrSinSala
	}
	cuerpo, _ := json.Marshal(dto.MensajeSaliente{
		Tipo: "reaccion", AudioID: audioID, Reaccion: reaccion, UsuarioID: origen.UsuarioID,
	})
	for c := range h.salas[audioID] {
		if c == origen {
			continue // quien reaccion ya vio su propia animación
		}
		if c.Enviar(cuerpo) {
			entregadas++
		}
	}
	return entregadas, audioID, nil
}
