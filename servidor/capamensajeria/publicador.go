package capamensajeria

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const NombreCola = "reproducciones"

// ReproduccionMensaje es el contrato JSON con ms-estadisticas
type ReproduccionMensaje struct {
	UsuarioID int       `json:"usuario_id"`
	AudioID   int       `json:"audio_id"`
	Fecha     time.Time `json:"fecha"`
}

type Publicador struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	mu   sync.Mutex // un solo canal compartido por varias goroutines
}

func NuevoPublicador(url string) (*Publicador, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	// durable=true: la cola sobrevive a un reinicio de RabbitMQ
	if _, err := ch.QueueDeclare(NombreCola, true, false, false, false, nil); err != nil {
		conn.Close()
		return nil, err
	}
	return &Publicador{conn: conn, ch: ch}, nil
}

// PublicarReproduccion envía el evento y no espera a que nadie lo consuma
// Si el publicador es nil (RabbitMQ no estaba disponible al arrancar) no hace nada:
// el streaming nunca debe fallar porque las estadísticas no estén disponibles
func (p *Publicador) PublicarReproduccion(usuarioID, audioID int) {
	if p == nil {
		return
	}
	cuerpo, _ := json.Marshal(ReproduccionMensaje{
		UsuarioID: usuarioID, AudioID: audioID, Fecha: time.Now(),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	p.mu.Lock()
	err := p.ch.PublishWithContext(ctx, "", NombreCola, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         cuerpo,
	})
	p.mu.Unlock()

	if err != nil {
		log.Printf("[COLA] error publicando: %v", err)
		return
	}
	fmt.Printf("[COLA] publicado -> usuario=%d audio=%d\n", usuarioID, audioID)
}

func (p *Publicador) Cerrar() {
	if p != nil {
		p.conn.Close()
	}
}
