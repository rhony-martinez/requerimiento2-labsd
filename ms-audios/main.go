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

	"ms-audios/controllers"
	"ms-audios/discovery"
	"ms-audios/middleware"
	"ms-audios/repositories"
	"ms-audios/services"
)

const nombreServicio = "ms-audios"

func main() {
	puerto, _ := strconv.Atoi(env("PORT", "8081"))
	host := env("SERVICE_HOST", "127.0.0.1")
	consulAddr := env("CONSUL_ADDR", "127.0.0.1:8500")
	secreto := env("GATEWAY_SECRET", "secreto-gateway-2026")

	repo := repositories.NewAudioRepository("datos/audios.json")
	almacen, err := repositories.NewAlmacenArchivos(env("DIRECTORIO_AUDIOS", "../servidor/canciones"))
	if err != nil {
		log.Fatalf("no fue posible preparar la carpeta de audios: %v", err)
	}
	service := services.NewAudioService(repo, almacen)
	controller := controllers.NewAudioController(service)

	router := gin.Default()
	router.Use(middleware.FiltroGateway(secreto))
	controller.RegistrarRutas(router, middleware.RequiereAdmin(env("ADMIN_TOKEN", "admin-2026")))

	srv := &http.Server{Addr: fmt.Sprintf(":%d", puerto), Handler: router}
	go func() {
		log.Printf("%s escuchando en :%d\n", nombreServicio, puerto)
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
