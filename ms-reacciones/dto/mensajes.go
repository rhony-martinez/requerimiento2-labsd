package dto

// MensajeEntrante es lo que envía un cliente.
type MensajeEntrante struct {
	Tipo     string `json:"tipo"` // "unirse" | "salir" | "reaccion"
	AudioID  int    `json:"audio_id,omitempty"`
	Reaccion string `json:"reaccion,omitempty"`
}

// MensajeSaliente es lo que recibe un cliente.
type MensajeSaliente struct {
	Tipo      string `json:"tipo"` // "reaccion" | "error"
	AudioID   int    `json:"audio_id,omitempty"`
	Reaccion  string `json:"reaccion,omitempty"`
	UsuarioID int    `json:"usuario_id,omitempty"`
	Error     string `json:"error,omitempty"`
}
