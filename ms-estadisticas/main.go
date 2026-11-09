package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"ms-estadisticas/controllers"
	"ms-estadisticas/discovery"
	"ms-estadisticas/messaging"
	"ms-estadisticas/middleware"
	"ms-estadisticas/repositories"
	"ms-estadisticas/services"
)

const nombreServicio = "ms-estadisticas"

func main() {
	puerto, _ := strconv.Atoi(env("PORT", "8084"))
	host := env("SERVICE_HOST", "127.0.0.1")
	consulAddr := env("CONSUL_ADDR", "127.0.0.1:8500")
	secreto := env("GATEWAY_SECRET", "secreto-gateway-2026")
	amqpURL := env("AMQP_URL", "amqp://guest:guest@127.0.0.1:5672/")

	repo, err := repositories.NewReproduccionRepository("datos/reproducciones.jsonl")
	if err != nil {
		log.Fatalf("no fue posible abrir el almacén de datos: %v", err)
	}
	defer repo.Cerrar()
	service := services.NewEstadisticasService(repo)
	controller := controllers.NewEstadisticasController(service)

	// Consumidor de la cola: corre en segundo plano, independiente del servidor HTTP
	consumidor := messaging.NuevoConsumidor(amqpURL, service)
	consumidor.Iniciar()

	router := gin.Default()
	router.Use(middleware.FiltroGateway(secreto))
	controller.RegistrarRutas(router)

	srv := &http.Server{Addr: fmt.Sprintf(":%d", puerto), Handler: router}
	go func() {
		log.Printf("%s escuchando en :%d", nombreServicio, puerto)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("error en el servidor: %v", err)
		}
	}()

	registro, err := discovery.Registrar(consulAddr, nombreServicio, host, puerto)
	if err != nil {
		log.Fatalf("no fue posible registrarse en Consul: %v", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	registro.Desregistrar()
	consumidor.Cerrar()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)

}

func env(clave, porDefecto string) string {
	if v := os.Getenv(clave); v != "" {
		return v
	}
	return porDefecto
}
