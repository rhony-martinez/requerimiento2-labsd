package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"ms-audios/dto"
	"ms-audios/entities"
	"ms-audios/repositories"
	"strings"
	"unicode/utf8"
)

type AudioService struct {
	repo    *repositories.AudioRepository
	almacen *repositories.AlmacenArchivos
}

func NewAudioService(repo *repositories.AudioRepository, almacen *repositories.AlmacenArchivos) *AudioService {
	return &AudioService{repo: repo, almacen: almacen}
}

// ErrValidacion se traduce a 400 en el controlador.
type ErrValidacion struct{ Mensaje string }

func (e *ErrValidacion) Error() string { return e.Mensaje }

type campoMeta struct{ clave, etiqueta string }

// Las claves coinciden con las de datos_iniciales.go. Se indexa por ID de tipo.
var camposPorTipo = map[int][]campoMeta{
	1: {{"artista_principal", "Artista principal"}, {"album", "Álbum"}, {"genero_musical", "Género musical"}},
	2: {{"nombre_podcast", "Nombre del podcast"}, {"titulo_episodio", "Título del episodio"}, {"anfitrion", "Anfitrión"}},
	3: {{"titulo_libro", "Título del libro"}, {"autor", "Autor"}, {"narrador", "Narrador"}},
	4: {{"tipo_sonido", "Tipo de sonido"}, {"fuente_audio", "Fuente del audio"}, {"uso_sugerido", "Uso sugerido"}},
}

const (
	formatoSoportado = "mp3"
	maxTitulo        = 150
	maxValor         = 200
	maxDuracionSeg   = 24 * 60 * 60
)

func (s *AudioService) ListarTipos() []dto.TipoAudioDTO {
	tipos := s.repo.ListarTipos()
	resultado := make([]dto.TipoAudioDTO, 0, len(tipos))
	for _, t := range tipos {
		resultado = append(resultado, dto.TipoAudioDTO{ID: t.ID, Nombre: t.Nombre})
	}
	return resultado
}

func (s *AudioService) ListarAudiosPorTipo(tipoID int) ([]dto.AudioResumenDTO, error) {
	audios, err := s.repo.ListarAudiosPorTipo(tipoID)
	if err != nil {
		return nil, err
	}
	resultado := make([]dto.AudioResumenDTO, 0, len(audios))
	for _, a := range audios {
		if a.TipoID == tipoID {
			resultado = append(resultado, dto.AudioResumenDTO{ID: a.ID, Titulo: a.Titulo})
		}
	}
	return resultado, nil
}

func (s *AudioService) ObtenerDetalle(id int) (dto.AudioDetalleDTO, error) {
	audio, err := s.repo.ObtenerAudio(id)
	if err != nil {
		return dto.AudioDetalleDTO{}, err
	}
	tipo, err := s.repo.ObtenerTipo(audio.TipoID)
	if err != nil {
		return dto.AudioDetalleDTO{}, err
	}
	return toDetalleDTO(audio, tipo.Nombre), nil
}

func toDetalleDTO(a entities.Audio, nombreTipo string) dto.AudioDetalleDTO {
	return dto.AudioDetalleDTO{
		ID: a.ID, Tipo: nombreTipo, Titulo: a.Titulo, Archivo: a.Archivo, Formato: a.Formato, DuracionSeg: a.DuracionSeg, Metadatos: a.Metadatos,
	}
}

func (s *AudioService) CamposDeTipo(tipoID int) ([]dto.CampoMetadatoDTO, error) {
	if _, err := s.repo.ObtenerTipo(tipoID); err != nil {
		return nil, err
	}
	resultado := make([]dto.CampoMetadatoDTO, 0)
	for _, c := range camposPorTipo[tipoID] {
		resultado = append(resultado, dto.CampoMetadatoDTO{Clave: c.clave, Etiqueta: c.etiqueta})
	}
	return resultado, nil
}

func (s *AudioService) RegistrarAudio(n dto.NuevoAudioDTO, contenido io.Reader) (dto.AudioDetalleDTO, error) {
	tipo, err := s.repo.ObtenerTipo(n.TipoID)
	if err != nil {
		return dto.AudioDetalleDTO{}, err
	}

	// 1. Validar TODO antes de tocar el disco.
	titulo := strings.TrimSpace(n.Titulo)
	if titulo == "" || utf8.RuneCountInString(titulo) > maxTitulo {
		return dto.AudioDetalleDTO{}, &ErrValidacion{fmt.Sprintf("el título es obligatorio (máximo %d caracteres)", maxTitulo)}
	}
	if n.DuracionSeg <= 0 || n.DuracionSeg > maxDuracionSeg {
		return dto.AudioDetalleDTO{}, &ErrValidacion{"la duración debe estar entre 1 segundo y 24 horas"}
	}
	metadatos, err := validarMetadatos(tipo.ID, n.Metadatos)
	if err != nil {
		return dto.AudioDetalleDTO{}, err
	}

	// 2. Guardar el archivo con un nombre generado por nosotros.
	archivo, err := nuevoNombreArchivo()
	if err != nil {
		return dto.AudioDetalleDTO{}, err
	}
	nombreEnDisco := archivo + "." + formatoSoportado
	if err := s.almacen.Guardar(nombreEnDisco, contenido); err != nil {
		return dto.AudioDetalleDTO{}, err
	}

	// 3. Registrar en el catálogo; si falla, no dejamos el archivo huérfano.
	audio, err := s.repo.Agregar(entities.Audio{
		TipoID: tipo.ID, Titulo: titulo, Archivo: archivo, Formato: formatoSoportado,
		DuracionSeg: n.DuracionSeg, Metadatos: metadatos,
	})
	if err != nil {
		s.almacen.Eliminar(nombreEnDisco)
		return dto.AudioDetalleDTO{}, err
	}
	return toDetalleDTO(audio, tipo.Nombre), nil
}

// validarMetadatos exige exactamente los campos del tipo e ignora cualquier otro.
func validarMetadatos(tipoID int, recibidos map[string]string) (map[string]string, error) {
	campos, ok := camposPorTipo[tipoID]
	if !ok {
		return nil, &ErrValidacion{"este tipo de audio no admite registro"}
	}
	limpios := make(map[string]string, len(campos))
	for _, c := range campos {
		valor := strings.TrimSpace(recibidos[c.clave])
		if valor == "" {
			return nil, &ErrValidacion{"falta el metadato obligatorio: " + c.etiqueta}
		}
		if utf8.RuneCountInString(valor) > maxValor {
			return nil, &ErrValidacion{fmt.Sprintf("%s es demasiado largo (máximo %d caracteres)", c.etiqueta, maxValor)}
		}
		limpios[c.clave] = valor
	}
	return limpios, nil
}

func nuevoNombreArchivo() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "audio_" + hex.EncodeToString(b), nil
}
