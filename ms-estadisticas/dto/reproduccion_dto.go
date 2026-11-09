package dto

import "time"

// Lo que llega por la cola
type ReproduccionMensajeDTO struct {
	UsuarioID int       `json:"usuario_id"`
	AudioID   int       `json:"audio_id"`
	Fecha     time.Time `json:"fecha"`
}

// Lo que se devuelve al cliente: un audio y cuántas veces lo reprodujo
type HistorialItemDTO struct {
	AudioID        int       `json:"audio_id"`
	Reproducciones int       `json:"reproducciones"`
	UltimaFecha    time.Time `json:"ultima_fecha"`
}
