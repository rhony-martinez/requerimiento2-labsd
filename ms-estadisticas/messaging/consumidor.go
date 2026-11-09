package messaging

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"ms-estadisticas/dto"
	"ms-estadisticas/services"
)

const NombreCola = "reproducciones"

type Consumidor struct {
	conn    *amqp.Connection
	ch      *amqp.Channel
	service *services.EstadisticasService
}

func NuevoConsumidor(url string, service *services.EstadisticasService) (*Consumidor, error) {
	var conn *amqp.Connection
	var err error
	// Reintentos: RabbitMQ puede tardar unos segundos en estar listo
	for intento := 1; intento <= 10; intento++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}
		log.Printf("RabbitMQ no disponible (intento %d/10): %v", intento, err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	if _, err := ch.QueueDeclare(NombreCola, true, false, false, false, nil); err != nil {
		conn.Close()
		return nil, err
	}
	// Prefetch: RabbitMQ entrega máximo 10 mensajes sin confirmar a la vez
	if err := ch.Qos(10, 0, false); err != nil {
		conn.Close()
		return nil, err
	}
	return &Consumidor{conn: conn, ch: ch, service: service}, nil
}

func (c *Consumidor) Iniciar() error {
	// autoAck=false: confirmamos a mano DESPUÉS de procesar. Si morimos a mitad,
	// RabbitMQ reentrega el mensaje en vez de perderlo.
	mensajes, err := c.ch.Consume(NombreCola, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for m := range mensajes {
			c.procesar(m)
		}
	}()
	fmt.Printf("[COLA] consumidor escuchando la cola %q\n", NombreCola)
	return nil
}

func (c *Consumidor) procesar(m amqp.Delivery) {
	var msg dto.ReproduccionMensajeDTO
	if err := json.Unmarshal(m.Body, &msg); err != nil {
		log.Printf("[COLA] mensaje ilegible, descartado: %v", err)
		m.Nack(false, false) // requeue=false: reintentarlo no lo arreglaría
		return
	}
	if err := c.service.RegistrarReproduccion(msg); err != nil {
		log.Printf("[COLA] mensaje inválido, descartado: %v", err)
		m.Nack(false, false)
		return
	}
	fmt.Printf("[COLA] consumido -> usuario=%d audio=%d fecha=%s\n",
		msg.UsuarioID, msg.AudioID, msg.Fecha.Format("2006-01-02 15:04:05"))
	m.Ack(false)
}

func (c *Consumidor) Cerrar() {
	c.conn.Close()
}
