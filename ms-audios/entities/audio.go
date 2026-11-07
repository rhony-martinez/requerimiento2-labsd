package entities

type TipoAudio struct {
	ID     int
	Nombre string
}

type Audio struct {
	ID          int
	TipoID      int
	Titulo      string // nombre para mostrar
	Archivo     string // nombre base del archivo en servidor/canciones (sin extensión)
	Formato     string // mp3, wav...
	DuracionSeg int
	Metadatos   map[string]string // campos propios de cada tipo
}
