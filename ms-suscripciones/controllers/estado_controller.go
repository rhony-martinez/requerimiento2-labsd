package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"pasarela-pagos/config"
)

// EstadoController expone la configuración de simulación activa, para que el
// cliente la muestre y quede visible en las capturas de pantalla.
type EstadoController struct {
	cfg config.Config
}

func NuevoEstadoController(cfg config.Config) *EstadoController {
	return &EstadoController{cfg: cfg}
}

func (ctl *EstadoController) Registrar(rg *gin.RouterGroup) {
	rg.GET("/estado", ctl.estado)
}

// GET /api/estado
func (ctl *EstadoController) estado(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"estado":           "activo",
		"modo_fallo":       ctl.cfg.ModoFallo,
		"cantidad_fallos":  ctl.cfg.CantidadFallos,
		"retardo_fallo_ms": ctl.cfg.RetardoFallo.Milliseconds(),
	})
}
