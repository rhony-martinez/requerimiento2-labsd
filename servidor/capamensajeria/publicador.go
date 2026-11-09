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
	url  string
	conn *amqp.Connection
	ch   *amqp.Channel
	mu   sync.Mutex // un solo canal compartido por varias goroutines
}

// NuevoPublicador ya no conecta ni falla: la conexión se establece al publicar.
func NuevoPublicador(url string) *Publicador {
	return &Publicador{url: url}
}

func (p *Publicador) asegurarCanal() error {
	if p.conn != nil && !p.conn.IsClosed() && p.ch != nil && !p.ch.IsClosed() {
		return nil
	}
	p.cerrarSinLock() // limpia resto de una conexión caída

	// Timeout de 3 s: si RabbitMQ no está, no dejamos colgado al streaming.
	conn, err := amqp.DialConfig(p.url, amqp.Config{Dial: amqp.DefaultDial(3 * time.Second)})
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return err
	}
	if _, err := ch.QueueDeclare(NombreCola, true, false, false, false, nil); err != nil {
		conn.Close()
		return err
	}
	p.conn, p.ch = conn, ch
	log.Println("[COLA] conectado a RabbitMQ")
	return nil
}

func (p *Publicador) cerrarSinLock() {
	if p.conn != nil {
		p.conn.Close()
	}
	p.conn, p.ch = nil, nil
}

func (p *Publicador) PublicarReproduccion(usuarioID, audioID int) {
	if p == nil {
		return
	}
	cuerpo, _ := json.Marshal(ReproduccionMensaje{
		UsuarioID: usuarioID, AudioID: audioID, Fecha: time.Now(),
	})

	p.mu.Lock()
	defer p.mu.Unlock()

	var err error
	// Dos intentos: si el primero falla por una conexión muerta,
	// el segundo reconecta desde cero.
	for intento := 1; intento <= 2; intento++ {
		if err = p.asegurarCanal(); err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			err = p.ch.PublishWithContext(ctx, "", NombreCola, false, false, amqp.Publishing{
				ContentType:  "application/json",
				DeliveryMode: amqp.Persistent,
				Body:         cuerpo,
			})
			cancel()
			if err == nil {
				fmt.Printf("[COLA] publicado -> usuario=%d audio=%d\n", usuarioID, audioID)
				return
			}
		}
		p.cerrarSinLock() // fuerza reconexión en el siguiente intento
	}
	log.Printf("[COLA] no se pudo publicar (usuario=%d audio=%d): %v", usuarioID, audioID, err)
}

func (p *Publicador) Cerrar() {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cerrarSinLock()
}
