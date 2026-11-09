package controllers

import (
	"fmt"
	"ms-estadisticas/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type EstadisticasController struct {
	service *services.EstadisticasService
}

func NewEstadisticasController(service *services.EstadisticasService) *EstadisticasController {
	return &EstadisticasController{service: service}
}

func (c *EstadisticasController) RegistrarRutas(r gin.IRouter) {
	r.GET("/estadisticas/usuarios/:id/historial", c.Historial)
}

func (c *EstadisticasController) Historial(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "el id del usuario debe ser un número positivo"})
		return
	}
	fmt.Printf("[REST] GET /estadisticas/usuarios/%d/historial -> consultando historial\n", id)
	ctx.JSON(http.StatusOK, c.service.HistorialDelUsuario(id))
}
