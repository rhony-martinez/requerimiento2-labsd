package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	capacontroladores "servidor.local/grpc-servidor/capaControladores"
	"servidor.local/grpc-servidor/capalogger"
	"servidor.local/grpc-servidor/capamensajeria"
	pb "servidor.local/grpc-servidor/serviciosCancion"
)

func main() {

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Error escuchando en el puerto: %v", err)
	}

	grpcServer := grpc.NewServer()

	// Crear logger como singleton
	objLogger := capalogger.CrearUnicaInstanciaDelLogger()
	defer objLogger.CerrarArchivo()

	// Si RabbitMQ no está, seguimos sin estadísticas en vez de impedir el streaming
	// ? Cola y credenciales
	publicador := capamensajeria.NuevoPublicador("amqp://guest:guest@127.0.0.1:5672/")
	defer publicador.Cerrar()

	objControlador := capacontroladores.NewControladorServidor(objLogger, publicador)

	// Se registra el controlador que ofrece el procedimiento remoto
	pb.RegisterAudioServiceServer(grpcServer, objControlador)

	fmt.Println("Servidor gRPC escuchando en :50051...")

	// Iniciar el servidor
	if errServ := grpcServer.Serve(listener); errServ != nil {
		log.Fatalf("Error al iniciar servidor gRPC: %v", errServ)
	}
}
