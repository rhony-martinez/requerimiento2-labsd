package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ms-audios/repositories"
	"ms-audios/services"
)

type AudioController struct {
	service *services.AudioService
}

func NewAudioController(service *services.AudioService) *AudioController {
	return &AudioController{service: service}
}

func (c *AudioController) RegistrarRutas(r gin.IRouter) {
	r.GET("/catalogo/tipos", c.ListarTipos)
	r.GET("/catalogo/tipo/:id/audios", c.ListarAudiosPorTipo)
	r.GET("/catalogo/audio/:id", c.ObtenerDetalle)
}

func (c *AudioController) ListarTipos(ctx *gin.Context) {
	fmt.Println("[REST] GET /catalogo/tipos -> consultando tipos de audio")
	ctx.JSON(http.StatusOK, c.service.ListarTipos())
}

func (c *AudioController) ListarAudiosPorTipo(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "el id del tipo debe ser numérico"})
		return
	}
	fmt.Printf("[REST] GET /catalogo/tipos/%d/audios -> listando audios del tipo\n", id)

	audios, err := c.service.ListarAudiosPorTipo(id)
	if errors.Is(err, repositories.ErrTipoNoEncontrado) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, audios)
}

func (c *AudioController) ObtenerDetalle(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "el id del audio debe ser numérico"})
		return
	}
	fmt.Printf("[REST] GET /catalogo/audio/%d -> consultando detalle del audio\n", id)

	detalle, err := c.service.ObtenerDetalle(id)
	if errors.Is(err, repositories.ErrAudioNoEncontrado) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, detalle)
}
