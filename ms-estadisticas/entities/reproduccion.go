package entities

import "time"

type Reproduccion struct {
	UsuarioID int       `json:"usuario_id"`
	AudioID   int       `json:"audio_id"`
	Fecha     time.Time `json:"fecha"`
}
