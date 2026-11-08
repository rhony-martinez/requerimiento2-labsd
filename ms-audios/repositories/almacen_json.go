package repositories

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"ms-audios/entities"
)

type contenidoJSON struct {
	Tipos  []entities.TipoAudio `json:"tipos"`
	Audios []entities.Audio     `json:"audios"`
}

// leerJSON devuelve os.ErrNotExist (envuelto) si el archivo no existe.
// Eso permite al llamado distinguir "primer arranque" de "archivo dañado"
func leerJSON(ruta string) (contenidoJSON, error) {
	var contenido contenidoJSON
	bytes, err := os.ReadFile(ruta)
	if err != nil {
		return contenido, err
	}
	if len(bytes) == 0 {
		return contenido, errors.New("el archivo está vacío")
	}
	err = json.Unmarshal(bytes, &contenido)
	return contenido, err
}

// escribirJSON crea la carpeta si falta y escribe de forma segura:
// Primero a un archivo temporal y luego lo renombra, para que un corte
// a mitad de escritura no deje el archivo real corrupto
func escribirJSON(ruta string, contenido contenidoJSON) error {
	if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
		return err
	}
	bytes, err := json.MarshalIndent(contenido, "", " ")
	if err != nil {
		return err
	}
	temporal := ruta + ".tmp"
	if err := os.WriteFile(temporal, bytes, 0o644); err != nil {
		return err
	}
	return os.Rename(temporal, ruta)
}
