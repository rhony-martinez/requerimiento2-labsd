package repositories

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"

	"ms-estadisticas/entities"
)

type ReproduccionRepository struct {
	mu      sync.RWMutex
	datos   []entities.Reproduccion
	archivo *os.File
}

func NewReproduccionRepository(ruta string) (*ReproduccionRepository, error) {
	if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
		return nil, err
	}
	datos, err := cargar(ruta)
	if err != nil {
		return nil, err
	}
	archivo, err := os.OpenFile(ruta, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	log.Printf("historial cargado: %d reproducciones desde %s", len(datos), ruta)
	return &ReproduccionRepository{datos: datos, archivo: archivo}, nil
}

// cargar lee el archivo línea a línea. Si no existe, devuelve una lista vacía.
func cargar(ruta string) ([]entities.Reproduccion, error) {
	datos := make([]entities.Reproduccion, 0)
	f, err := os.Open(ruta)
	if os.IsNotExist(err) {
		return datos, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for n := 1; scanner.Scan(); n++ {
		var rep entities.Reproduccion
		if err := json.Unmarshal(scanner.Bytes(), &rep); err != nil {
			// Típico de una línea cortada por un apagón: se salta, no se aborta.
			log.Printf("línea %d inválida, se omite: %v", n, err)
			continue
		}
		datos = append(datos, rep)
	}
	return datos, scanner.Err()
}

func (r *ReproduccionRepository) Guardar(rep entities.Reproduccion) error {
	linea, err := json.Marshal(rep)
	if err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	// Primero el disco, luego la memoria: si falla la escritura, no queda
	// en memoria algo que se perdería al reiniciar.
	if _, err := r.archivo.Write(append(linea, '\n')); err != nil {
		return err
	}
	r.datos = append(r.datos, rep)
	return nil
}

func (r *ReproduccionRepository) ListarPorUsuario(usuarioID int) []entities.Reproduccion {
	r.mu.RLock()
	defer r.mu.RUnlock()
	resultado := make([]entities.Reproduccion, 0)
	for _, rep := range r.datos {
		if rep.UsuarioID == usuarioID {
			resultado = append(resultado, rep)
		}
	}
	return resultado
}

func (r *ReproduccionRepository) Cerrar() {
	r.archivo.Close()
}
