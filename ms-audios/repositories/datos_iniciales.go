package repositories

import "ms-audios/entities"

func (r *AudioRepository) cargarDatosIniciales() {
	r.tipos = []entities.TipoAudio{
		{ID: 1, Nombre: "Música"},
		{ID: 2, Nombre: "Podcasts"},
		{ID: 3, Nombre: "Audiolibros"},
		{ID: 4, Nombre: "Ruido Blanco"},
	}

	r.audios = []entities.Audio{
		{
			ID:          1,
			TipoID:      1,
			Titulo:      "Lamento boliviano",
			Archivo:     "cancion1",
			Formato:     "mp3",
			DuracionSeg: 258,
			Metadatos: map[string]string{
				"artista_principal": "Enanitos verdes",
				"album":             "Contrabando",
				"genero_musical":    "Rock",
			},
		},
		{
			ID:          2,
			TipoID:      1,
			Titulo:      "Segunda canción",
			Archivo:     "cancion2",
			Formato:     "mp3",
			DuracionSeg: 200,
			Metadatos: map[string]string{
				"artista_principal": "Artista de prueba",
				"album":             "Álbum de prueba",
				"genero_musical":    "Pop",
			},
		},
		{
			ID:          3,
			TipoID:      2,
			Titulo:      "Primer Podcast",
			Archivo:     "podcast1",
			Formato:     "mp3",
			DuracionSeg: 258,
			Metadatos: map[string]string{
				"nombre_podcast":  "Podcast 1 de prueba",
				"titulo_episodio": "Episodio de prueba",
				"anfitrion":       "Anfitrion de podcast 1",
			},
		},
		{
			ID:          4,
			TipoID:      2,
			Titulo:      "Segundo Podcast",
			Archivo:     "podcast2",
			Formato:     "mp3",
			DuracionSeg: 200,
			Metadatos: map[string]string{
				"nombre_podcast":  "Podcast 2 de prueba",
				"titulo_episodio": "Episodio de prueba",
				"anfitrion":       "Anfitrion de podcast 2",
			},
		},
		{
			ID:          5,
			TipoID:      3,
			Titulo:      "Primer Audiolibro",
			Archivo:     "libro1",
			Formato:     "mp3",
			DuracionSeg: 258,
			Metadatos: map[string]string{
				"titulo_libro": "Libro 1 de prueba",
				"autor":        "Autor 1",
				"narrador":     "Narrador 1",
			},
		},
		{
			ID:          6,
			TipoID:      3,
			Titulo:      "Segundo Audiolibro",
			Archivo:     "libro2",
			Formato:     "mp3",
			DuracionSeg: 200,
			Metadatos: map[string]string{
				"titulo_libro": "Libro 2 de prueba",
				"autor":        "Autor 2",
				"narrador":     "Narrador 2",
			},
		},
		{
			ID:          7,
			TipoID:      4,
			Titulo:      "Primer Ruido Blanco",
			Archivo:     "ruido1",
			Formato:     "mp3",
			DuracionSeg: 258,
			Metadatos: map[string]string{
				"tipo_sonido":  "Sonido 1",
				"fuente_audio": "Fuente 1",
				"uso_sugerido": "Uso sugerido 1",
			},
		},
		{
			ID:          8,
			TipoID:      4,
			Titulo:      "Segundo Ruido Blanco",
			Archivo:     "ruido2",
			Formato:     "mp3",
			DuracionSeg: 200,
			Metadatos: map[string]string{
				"tipo_sonido":  "Sonido 2",
				"fuente_audio": "Fuente 2",
				"uso_sugerido": "Uso sugerido 2",
			},
		},
	}
}
