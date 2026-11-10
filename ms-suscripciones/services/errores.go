package services

import "net/http"

// ErrorServicio es un error de negocio con el código HTTP que el controlador debe devolver.
type ErrorServicio struct {
	Codigo  int
	Mensaje string
}

func (e *ErrorServicio) Error() string { return e.Mensaje }

func errorValidacion(mensaje string) *ErrorServicio {
	return &ErrorServicio{Codigo: http.StatusBadRequest, Mensaje: mensaje}
}

func errorNoEncontrado(mensaje string) *ErrorServicio {
	return &ErrorServicio{Codigo: http.StatusNotFound, Mensaje: mensaje}
}

func errorNegocio(mensaje string) *ErrorServicio {
	return &ErrorServicio{Codigo: http.StatusUnprocessableEntity, Mensaje: mensaje}
}

func errorInterno(mensaje string) *ErrorServicio {
	return &ErrorServicio{Codigo: http.StatusInternalServerError, Mensaje: mensaje}
}
