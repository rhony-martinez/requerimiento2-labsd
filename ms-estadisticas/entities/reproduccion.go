package entities

import "time"

type Reproduccion struct {
	UsuarioID int
	AudioID   int
	Fecha     time.Time
}
