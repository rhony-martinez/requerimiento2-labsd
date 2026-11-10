package dto

type TipoAudioDTO struct {
	ID     int    `json:"id"`
	Nombre string `json:"nombre"`
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

// Un campo de metadato propio de un tipo (para que el cliente arme el formulario).
type CampoMetadatoDTO struct {
	Clave    string `json:"clave"`
	Etiqueta string `json:"etiqueta"`
}

// Datos del formulario multipart. El archivo viaja aparte, como io.Reader.
type NuevoAudioDTO struct {
	TipoID      int
	Titulo      string
	DuracionSeg int
	Metadatos   map[string]string
}
