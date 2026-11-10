package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"pasarela-pagos/services"
	"pasarela-pagos/services/dto"
)

// responderError traduce los errores de la fachada a respuestas HTTP.
func responderError(c *gin.Context, err error) {
	var errServicio *services.ErrorServicio
	if errors.As(err, &errServicio) {
		c.JSON(errServicio.Codigo, dto.ErrorDTO{Error: errServicio.Mensaje})
		return
	}
	c.JSON(http.StatusInternalServerError, dto.ErrorDTO{Error: err.Error()})
}

func responderCuerpoInvalido(c *gin.Context) {
	c.JSON(http.StatusBadRequest, dto.ErrorDTO{Error: "el cuerpo de la petición no es un JSON válido"})
}
