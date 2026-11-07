package dto

type TipoAudioDTO struct {
	ID     int    `json:"id"`
	Nombre string `json:"titulo"`
}

// Para la lista: solo lo necesario para pintar el menú
type AudioResumenDTO struct {
	ID     int    `json:"id"`
	Titulo string `json:"titulo"`
}

// Para el detalle: incluye el nombre del tipo y los metadatos
type AudioDetalleDTO struct {
	ID          int               `json:"id"`
	Tipo        string            `json:"tipo"`
	Titulo      string            `json:"titulo"`
	Archivo     string            `json:"archivo"`
	Formato     string            `json:"formato"`
	DuracionSeg int               `json:"duracion_seg"`
	Metadatos   map[string]string `json:"metadatos"`
}
