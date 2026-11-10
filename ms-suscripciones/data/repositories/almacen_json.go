package repositories

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Los repositorios guardan sus datos en archivos JSON para que el historial
// sobreviva a los reinicios del servidor entre un caso de prueba y otro.

func cargarJSON(ruta string, destino any) error {
	contenido, err := os.ReadFile(ruta)
	if os.IsNotExist(err) || (err == nil && len(contenido) == 0) {
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(contenido, destino)
}

func guardarJSON(ruta string, datos any) error {
	if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
		return err
	}
	contenido, err := json.MarshalIndent(datos, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ruta, contenido, 0o644)
}
