package entities

type TipoAudio struct {
	ID     int    `json:"id"`
	Nombre string `json:"nombre"`
}

type Audio struct {
	ID          int               `json:"id"`
	TipoID      int               `json:"tipo_id"`
	Titulo      string            `json:"titulo"`  // nombre para mostrar
	Archivo     string            `json:"archivo"` // nombre base del archivo en servidor/canciones (sin extensión)
	Formato     string            `json:"formato"` // mp3, wav...
	DuracionSeg int               `json:"duracion_seg"`
	Metadatos   map[string]string `json:"metadatos"` // campos propios de cada tipo
}
