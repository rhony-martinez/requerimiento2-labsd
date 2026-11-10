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

	temporal := ruta + ".tmp"
	f, err := os.OpenFile(temporal, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(contenido); err != nil {
		f.Close()
		os.Remove(temporal)
		return err
	}
	// Sync: pide al sistema operativo que baje los datos al disco ahora.
	// Sin esto, el renombrado podría quedar guardado y el contenido no (corte de luz).
	if err := f.Sync(); err != nil {
		f.Close()
		os.Remove(temporal)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(temporal)
		return err
	}
	return os.Rename(temporal, ruta)
}
