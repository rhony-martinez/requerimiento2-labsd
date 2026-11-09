package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"ms-reacciones/dto"
	"ms-reacciones/services"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = 30 * time.Second
)

// Aquí CheckOrigin permite todo porque este servicio solo es alcanzable a través
// del gateway (FiltroGateway). Origin lo controlaría un servicio expuesto directamente.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type ReaccionesController struct {
	hub *services.Hub
}

func NewReaccionesController(hub *services.Hub) *ReaccionesController {
	return &ReaccionesController{hub: hub}
}

func (c *ReaccionesController) RegistrarRutas(r gin.IRouter) {
	r.GET("/reacciones/ws", c.Conectar)
}

func (c *ReaccionesController) Conectar(ctx *gin.Context) {
	usuarioID, err := strconv.Atoi(ctx.Query("usuario_id"))
	if err != nil || usuarioID <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "usuario_id debe ser un número positivo"})
		return
	}
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		fmt.Printf("[WS] fallo en el handshake: %v\n", err)
	}

	cliente := services.NuevoCliente(usuarioID)
	fmt.Printf("[WS] conectado    usuario=%d\n", usuarioID)

	go c.escribir(conn, cliente)
	c.leer(conn, cliente) // bloquea hasta que el lciente se vaya

	// Limpieza: sin esto el hub seguiría enviando a un cliente muerto.
	c.hub.Salir(cliente)
	close(cliente.Cerrado)
	conn.Close()
	fmt.Printf("[WS] desconectado usuario=%d\n", usuarioID)
}

// leer procesa los mensajes del cliente. Si devuelve, la conexión terminó
func (c *ReaccionesController) leer(conn *websocket.Conn, cliente *services.Cliente) {
	conn.SetReadLimit(1024) // evitamos que se manden mensajes enormes (seguridad)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, bytes, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var m dto.MensajeEntrante
		if err := json.Unmarshal(bytes, &m); err != nil {
			responderError(cliente, "mensaje no es JSON válido")
			continue
		}

		switch m.Tipo {
		case "unirse":
			if err := c.hub.Unir(cliente, m.AudioID); err != nil {
				responderError(cliente, err.Error())
				continue
			}
			fmt.Printf("[WS] usuario=%d se une a la sala del audio %d\n", cliente.UsuarioID, m.AudioID)

		case "salir":
			c.hub.Salir(cliente)
			fmt.Printf("[WS] usuario=%d sale de su sala\n", cliente.UsuarioID)

		case "reaccion":
			n, audioID, err := c.hub.Reacciones(cliente, m.Reaccion)
			if err != nil {
				responderError(cliente, err.Error())
				continue
			}
			fmt.Printf("[WS] reacción %q de usuario=%d en audio %d -> reenviada a %d cliente(s)\n",
				m.Reaccion, cliente.UsuarioID, audioID, n)

		default:
			responderError(cliente, "tipo de mensaje desconocido: "+m.Tipo)
		}

	}
}

// escribir es la única goroutine que escribe en la conexión
func (c *ReaccionesController) escribir(conn *websocket.Conn, cliente *services.Cliente) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	for {
		select {
		case msg := <-cliente.Envio:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				conn.Close() // despierta a leer()
				return
			}
		case <-ticker.C:
			// El ping deteca conexiones muertas sin que el cliente haga nada
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				conn.Close()
				return
			}
		case <-cliente.Cerrado:
			return
		}
	}
}

func responderError(cliente *services.Cliente, texto string) {
	cuerpo, _ := json.Marshal(dto.MensajeSaliente{Tipo: "error", Error: texto})
	cliente.Enviar(cuerpo)
}
