package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"pasarela-pagos/services"
	"pasarela-pagos/services/dto"
)

type PagoController struct {
	servicio        *services.PagoService
	simuladorFallos gin.HandlerFunc
}

// NuevoPagoController recibe el middleware de simulación de fallos, que solo se
// aplica al servicio de realizar pagos.
func NuevoPagoController(servicio *services.PagoService, simuladorFallos gin.HandlerFunc) *PagoController {
	return &PagoController{servicio: servicio, simuladorFallos: simuladorFallos}
}

func (ctl *PagoController) Registrar(rg *gin.RouterGroup) {
	rg.POST("/tokens", ctl.generarToken)
	rg.POST("/pagos", ctl.simuladorFallos, ctl.realizarPago)
	rg.GET("/transacciones/:cedula", ctl.consultarTransacciones)
}

// POST /suscripciones/tokens — Servicio REST 1: generar token de idempotencia.
func (ctl *PagoController) generarToken(c *gin.Context) {
	var sol dto.SolicitudTokenDTO
	if err := c.ShouldBindJSON(&sol); err != nil {
		responderCuerpoInvalido(c)
		return
	}
	fmt.Printf("[REST] POST /suscripciones/tokens -> cédula=%s plan=%s\n", sol.Cedula, sol.SuscripcionID)
	respuesta, err := ctl.servicio.GenerarToken(sol)
	if err != nil {
		responderError(c, err)
		return
	}
	c.JSON(http.StatusCreated, respuesta)
}

// POST /api/pagos — Servicio REST 2: realizar pago.
func (ctl *PagoController) realizarPago(c *gin.Context) {
	var sol dto.SolicitudPagoDTO
	if err := c.ShouldBindJSON(&sol); err != nil {
		responderCuerpoInvalido(c)
		return
	}
	fmt.Printf("[REST] POST /suscripciones/pagos -> token=%s monto=%.0f\n", sol.TokenIdempotencia, sol.Monto)
	respuesta, err := ctl.servicio.RealizarPago(sol)
	if err != nil {
		responderError(c, err)
		return
	}
	c.JSON(http.StatusOK, respuesta)
}

// GET /api/transacciones/:cedula — Servicio REST 3: consultar transacciones.
func (ctl *PagoController) consultarTransacciones(c *gin.Context) {
	respuesta, err := ctl.servicio.ConsultarTransacciones(c.Param("cedula"))
	if err != nil {
		responderError(c, err)
		return
	}
	c.JSON(http.StatusOK, respuesta)
}
