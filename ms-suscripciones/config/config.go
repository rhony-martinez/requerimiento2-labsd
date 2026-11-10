package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// Modos de simulación de fallos soportados por el servidor.
const (
	ModoNinguno   = "NINGUNO"   // el servidor funciona normalmente
	ModoPeticion  = "PETICION"  // se pierden los mensajes de petición (el pago NO se procesa)
	ModoRespuesta = "RESPUESTA" // se pierden los mensajes de respuesta (el pago SÍ se procesa)
)

// Config agrupa la configuración del servidor leída desde variables de entorno.
type Config struct {
	Puerto          string
	ModoFallo       string
	CantidadFallos  int
	RetardoFallo    time.Duration
	DirectorioDatos string
	Secreto         string
	Host            string
	ConsulAddr      string
}

// Cargar lee las variables de entorno y aplica valores por defecto.
//
//	PUERTO            puerto HTTP (por defecto 8083)
//	MODO_FALLO        NINGUNO | PETICION | RESPUESTA (por defecto NINGUNO)
//	CANTIDAD_FALLOS   número de peticiones de pago, por token, que fallarán (por defecto 0)
//	RETARDO_FALLO_MS  tiempo que el servidor "cuelga" la conexión antes de cerrarla (por defecto 5000)
//	DIRECTORIO_DATOS  carpeta donde se persisten tokens y transacciones (por defecto ./datos)
func Cargar() Config {
	cfg := Config{
		Puerto:          valorOPorDefecto("PUERTO", "8083"),
		ModoFallo:       strings.ToUpper(valorOPorDefecto("MODO_FALLO", ModoNinguno)),
		CantidadFallos:  enteroOPorDefecto("CANTIDAD_FALLOS", 0),
		RetardoFallo:    time.Duration(enteroOPorDefecto("RETARDO_FALLO_MS", 5000)) * time.Millisecond,
		DirectorioDatos: valorOPorDefecto("DIRECTORIO_DATOS", "./datos"),
		Secreto:         valorOPorDefecto("GATEWAY_SECRET", "secreto-gateway-2026"),
		Host:            valorOPorDefecto("SERVICE_HOST", "127.0.0.1"),
		ConsulAddr:      valorOPorDefecto("CONSUL_ADDR", "127.0.0.1:8500"),
	}

	switch cfg.ModoFallo {
	case ModoNinguno, ModoPeticion, ModoRespuesta:
	default:
		log.Printf("MODO_FALLO '%s' no válido, se usará %s", cfg.ModoFallo, ModoNinguno)
		cfg.ModoFallo = ModoNinguno
	}
	if cfg.ModoFallo == ModoNinguno {
		cfg.CantidadFallos = 0
	}
	return cfg
}

func valorOPorDefecto(clave, porDefecto string) string {
	if v := strings.TrimSpace(os.Getenv(clave)); v != "" {
		return v
	}
	return porDefecto
}

func enteroOPorDefecto(clave string, porDefecto int) int {
	v := strings.TrimSpace(os.Getenv(clave))
	if v == "" {
		return porDefecto
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		log.Printf("Valor inválido para %s='%s', se usará %d", clave, v, porDefecto)
		return porDefecto
	}
	return n
}
