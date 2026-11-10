package repositories

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/go-mp3"
)

var (
	ErrNoEsMP3      = errors.New("el archivo no es un mp3 válido")
	ErrArchivoVacio = errors.New("el archivo está vacío")
)

type AlmacenArchivos struct {
	dir string
}

func NewAlmacenArchivos(dir string) (*AlmacenArchivos, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &AlmacenArchivos{dir: dir}, nil
}

// esMP3 reconoce una etiqueta ID3 ("ID3") o una cabecera de frame MPEG (11 bits en 1).
func esMP3(c []byte) bool {
	if len(c) < 3 {
		return false
	}
	if c[0] == 'I' && c[1] == 'D' && c[2] == '3' {
		return true
	}
	return c[0] == 0xFF && c[1]&0xE0 == 0xE0
}

// Guardar escribe el contenido en <dir>/<nombre>, de forma atómica.
func (a *AlmacenArchivos) Guardar(nombre string, contenido io.Reader) error {
	nombre = filepath.Base(nombre) // defensa: jamás una ruta
	destino := filepath.Join(a.dir, nombre)
	if _, err := os.Stat(destino); err == nil {
		return fmt.Errorf("el archivo %s ya existe", nombre)
	}

	lector := bufio.NewReader(contenido)
	cabecera, err := lector.Peek(3) // mira sin consumir
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	if len(cabecera) == 0 {
		return ErrArchivoVacio
	}
	if !esMP3(cabecera) {
		return ErrNoEsMP3
	}

	temporal := destino + ".tmp"
	f, err := os.OpenFile(temporal, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, lector); err != nil {
		f.Close()
		os.Remove(temporal)
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		os.Remove(temporal)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(temporal)
		return err
	}
	return os.Rename(temporal, destino)
}

func (a *AlmacenArchivos) Eliminar(nombre string) error {
	err := os.Remove(filepath.Join(a.dir, filepath.Base(nombre)))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// DuracionSegundos lee el mp3 ya guardado y calcula su duración real.
func (a *AlmacenArchivos) DuracionSegundos(nombre string) (int, error) {
	f, err := os.Open(filepath.Join(a.dir, filepath.Base(nombre)))
	if err != nil {
		return 0, err
	}
	defer f.Close()

	d, err := mp3.NewDecoder(f)
	if err != nil {
		return 0, ErrNoEsMP3 // pasó la firma inicial pero no es un mp3 decodificable
	}
	if d.Length() <= 0 || d.SampleRate() <= 0 {
		return 0, ErrNoEsMP3
	}
	// El decodificador entrega PCM de 16 bits y 2 canales: 4 bytes por muestra.
	seg := float64(d.Length()) / float64(d.SampleRate()*4)
	return int(math.Round(seg)), nil
}
