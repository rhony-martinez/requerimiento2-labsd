package messaging

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"ms-estadisticas/dto"
	"ms-estadisticas/services"
)

const NombreCola = "reproducciones"

type Consumidor struct {
	url     string
	service *services.EstadisticasService
	cerrado atomic.Bool
	mu      sync.Mutex
	conn    *amqp.Connection // la sesión actual, para poder cerrarla al apagar
}

func NuevoConsumidor(url string, service *services.EstadisticasService) *Consumidor {
	return &Consumidor{url: url, service: service}
}

// Iniciar lanza el bucle en segundo plano y vuelve de inmediato.
func (c *Consumidor) Iniciar() {
	go c.bucle()
}

func (c *Consumidor) bucle() {
	for !c.cerrado.Load() {
		err := c.sesion()
		if c.cerrado.Load() {
			return
		}
		log.Printf("[COLA] sesión terminada (%v); reintentando en 3 s", err)
		time.Sleep(3 * time.Second)
	}
}

// sesion abre una conexión y consume hasta que se cae. Devuelve por qué terminó.
func (c *Consumidor) sesion() error {
	conn, err := amqp.Dial(c.url)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(NombreCola, true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.Qos(10, 0, false); err != nil {
		return err
	}
	mensajes, err := ch.Consume(NombreCola, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()
	fmt.Printf("[COLA] consumidor escuchando la cola %q\n", NombreCola)

	// El canal "mensajes" se cierra solo cuando la conexión muere.
	for m := range mensajes {
		c.procesar(m)
	}
	return errors.New("canal de mensajes cerrado")
}

func (c *Consumidor) procesar(m amqp.Delivery) {
	var msg dto.ReproduccionMensajeDTO
	if err := json.Unmarshal(m.Body, &msg); err != nil {
		log.Printf("[COLA] mensaje ilegible, descartado: %v", err)
		m.Nack(false, false)
		return
	}

	err := c.service.RegistrarReproduccion(msg)
	switch {
	case errors.Is(err, services.ErrMensajeInvalido):
		// Reintentar no lo arregla: se descarta.
		log.Printf("[COLA] mensaje inválido, descartado: %v", err)
		m.Nack(false, false)
		return
	case err != nil:
		// Fallo nuestro (p. ej. disco): el mensaje es válido, se devuelve a la cola.
		log.Printf("[COLA] error guardando, se reencola: %v", err)
		m.Nack(false, true)
		return
	}

	fmt.Printf("[COLA] consumido -> usuario=%d audio=%d fecha=%s\n",
		msg.UsuarioID, msg.AudioID, msg.Fecha.Format("2006-01-02 15:04:05"))
	if err := m.Ack(false); err != nil {
		// Se guardó pero no se pudo confirmar: RabbitMQ lo reentregará (duplicado).
		log.Printf("[COLA] no se pudo confirmar: %v", err)
	}
}

func (c *Consumidor) Cerrar() {
	c.cerrado.Store(true)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		c.conn.Close()
	}
}
