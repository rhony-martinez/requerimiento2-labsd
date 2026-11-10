package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"pasarela-pagos/config"
	"pasarela-pagos/controllers"
	"pasarela-pagos/data/repositories"
	"pasarela-pagos/discovery"
	"pasarela-pagos/middleware"
	"pasarela-pagos/services"
)

const nombreServicio = "ms-suscripciones"

func main() {
	cfg := config.Cargar()

	// Capa de acceso a datos
	suscripcionRepo := repositories.NuevoSuscripcionRepository()
	tokenRepo, err := repositories.NuevoTokenRepository(filepath.Join(cfg.DirectorioDatos, "tokens.json"))
	if err != nil {
		log.Fatalf("No se pudo cargar el repositorio de tokens: %v", err)
	}
	transaccionRepo, err := repositories.NuevoTransaccionRepository(filepath.Join(cfg.DirectorioDatos, "transacciones.json"))
	if err != nil {
		log.Fatalf("No se pudo cargar el repositorio de transacciones: %v", err)
	}

	// Capa de fachada
	suscripcionService := services.NuevoSuscripcionService(suscripcionRepo)
	pagoService := services.NuevoPagoService(tokenRepo, transaccionRepo, suscripcionRepo)

	// Capa de controladores
	simulador := middleware.NuevoSimuladorFallos(cfg)
	router := gin.Default()
	router.Use(middleware.FiltroGateway(cfg.Secreto))
	api := router.Group("/suscripciones")
	controllers.NuevoSuscripcionController(suscripcionService).Registrar(api)
	controllers.NuevoPagoController(pagoService, simulador.Middleware()).Registrar(api)
	controllers.NuevoEstadoController(cfg).Registrar(api)

	puertoInt, _ := strconv.Atoi(cfg.Puerto)
	registro, err := discovery.Registrar(cfg.ConsulAddr, nombreServicio, cfg.Host, puertoInt)
	if err != nil {
		log.Fatalf("no fue posible registrarse en Consul: %v", err)
	}

	srv := &http.Server{Addr: fmt.Sprintf(":%d", puertoInt), Handler: router}
	go func() {
		log.Println("==============================================")
		log.Println(" Pasarela de pagos de suscripciones Spotify")
		log.Printf(" Puerto:            %s", cfg.Puerto)
		log.Printf(" MODO_FALLO:        %s", cfg.ModoFallo)
		log.Printf(" CANTIDAD_FALLOS:   %d", cfg.CantidadFallos)
		log.Printf(" RETARDO_FALLO_MS:  %d", cfg.RetardoFallo.Milliseconds())
		log.Printf(" DIRECTORIO_DATOS:  %s", cfg.DirectorioDatos)
		log.Println("==============================================")
		log.Printf("%s escuchando en :%d\n", nombreServicio, puertoInt)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("error en el servidor: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	registro.Desregistrar()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
