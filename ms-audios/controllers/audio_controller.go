package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ms-audios/dto"
	"ms-audios/repositories"
	"ms-audios/services"
)

const maxSubidaBytes = 50 << 20 // 50 MB

type AudioController struct {
	service *services.AudioService
}

func NewAudioController(service *services.AudioService) *AudioController {
	return &AudioController{service: service}
}

func (c *AudioController) RegistrarRutas(r gin.IRouter, admin gin.HandlerFunc) {
	r.GET("/catalogo/tipos", c.ListarTipos)
	r.GET("/catalogo/tipos/:id/audios", c.ListarAudiosPorTipo)
	r.GET("/catalogo/tipos/:id/campos", c.CamposDeTipo)
	r.GET("/catalogo/audios/:id", c.ObtenerDetalle)
	r.POST("/catalogo/audios", admin, c.Registrar) // solo administradores
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
	fmt.Printf("[REST] GET /catalogo/audios/%d -> consultando detalle del audio\n", id)

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

func (c *AudioController) CamposDeTipo(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "el id del tipo debe ser numérico"})
		return
	}
	fmt.Printf("[REST] GET /catalogo/tipos/%d/campos -> consultando metadatos del tipo\n", id)
	campos, err := c.service.CamposDeTipo(id)
	if errors.Is(err, repositories.ErrTipoNoEncontrado) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, campos)
}

func (c *AudioController) Registrar(ctx *gin.Context) {
	// Tope duro: si el cuerpo lo excede, la lectura falla.
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxSubidaBytes+(1<<20))
	// Hasta 8 MB en RAM; lo que exceda pasa a un archivo temporal.
	if err := ctx.Request.ParseMultipartForm(8 << 20); err != nil {
		var excedido *http.MaxBytesError
		if errors.As(err, &excedido) {
			ctx.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "el archivo supera el máximo de 50 MB"})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "se esperaba un formulario multipart/form-data"})
		return
	}

	tipoID, err := strconv.Atoi(ctx.Request.PostFormValue("tipo_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "tipo_id debe ser numérico"})
		return
	}
	duracion, _ := strconv.Atoi(ctx.Request.PostFormValue("duracion_seg")) // 0 si no es número: lo rechaza la fachada
	var metadatos map[string]string
	if err := json.Unmarshal([]byte(ctx.Request.PostFormValue("metadatos")), &metadatos); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "metadatos debe ser un objeto JSON"})
		return
	}
	archivo, cabecera, err := ctx.Request.FormFile("archivo")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "falta el archivo (campo \"archivo\")"})
		return
	}
	defer archivo.Close()
	if cabecera.Size > maxSubidaBytes {
		ctx.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "el archivo supera el máximo de 50 MB"})
		return
	}

	titulo := ctx.Request.PostFormValue("titulo")
	fmt.Printf("[REST] POST /catalogo/audios -> registrando %q (tipo=%d, %d bytes)\n", titulo, tipoID, cabecera.Size)

	detalle, err := c.service.RegistrarAudio(dto.NuevoAudioDTO{
		TipoID: tipoID, Titulo: titulo, DuracionSeg: duracion, Metadatos: metadatos,
	}, archivo)

	var errVal *services.ErrValidacion
	switch {
	case err == nil:
		fmt.Printf("[REST] audio registrado -> id=%d archivo=%s.%s\n", detalle.ID, detalle.Archivo, detalle.Formato)
		ctx.JSON(http.StatusCreated, detalle)
	case errors.As(err, &errVal):
		ctx.JSON(http.StatusBadRequest, gin.H{"error": errVal.Mensaje})
	case errors.Is(err, repositories.ErrTipoNoEncontrado):
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, repositories.ErrTituloDuplicado):
		ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, repositories.ErrNoEsMP3):
		ctx.JSON(http.StatusUnsupportedMediaType, gin.H{"error": err.Error()})
	case errors.Is(err, repositories.ErrArchivoVacio):
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		log.Printf("[REST] error interno registrando audio: %v", err) // el detalle, solo al log
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error interno al registrar el audio"})
	}
}
