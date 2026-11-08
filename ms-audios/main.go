package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"ms-audios/controllers"
	"ms-audios/repositories"
	"ms-audios/services"
)

func main() {
	puerto := env("PORT", "8081")

	repo := repositories.NewAudioRepository("datos/audios.json")
	service := services.NewAudioService(repo)
	controller := controllers.NewAudioController(service)

	router := gin.Default()
	controller.RegistrarRutas(router)

	fmt.Printf("ms-audios escuchando en :%s\n", puerto)
	if err := router.Run(":" + puerto); err != nil {
		log.Fatalf("error al iniciar el servidor: %v", err)
	}
}

func env(clave, porDefecto string) string {
	if v := os.Getenv(clave); v != "" {
		return v
	}
	return porDefecto
}
