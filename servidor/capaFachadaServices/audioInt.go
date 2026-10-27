package capafachadaservices

import (
	"fmt"
	"io"
	"log"
	"os"
)

func StreamAudioFile(tituloCancion string, funcionParaEnviarFragmento func([]byte) error) error {
	log.Printf("Canción solicitada: %s", tituloCancion)
	file, err := os.Open("canciones/" + tituloCancion)
	if err != nil {
		return fmt.Errorf("no se pudo abrir el archivo: %w", err)
	}
	defer file.Close()

	buffer := make([]byte, 64*1024) // 64 KB se envian por fragmento
	fragmento := 0

	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			log.Println("Canción enviada completamente desde la fachada.")
			break
		}
		if err != nil {
			return fmt.Errorf("error leyendo el archivo: %w", err)
		}

		fragmento++
		log.Printf("Fragmento #%d leído (%d bytes) y enviando", fragmento, n)
		//time.Sleep(1 * time.Second) // Simula un retardo en la lectura
		// ejecutamos la función para enviar el fragmento al cliente
		err = funcionParaEnviarFragmento(buffer[:n])
		if err != nil {
			return fmt.Errorf("error enviando fragmento #%d: %w", fragmento, err)
		}
	}

	return nil
}
