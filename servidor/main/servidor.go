package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	capacontroladores "servidor.local/grpc-servidor/capaControladores"
	"servidor.local/grpc-servidor/capalogger"
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

	objControlador := capacontroladores.NewControladorServidor(objLogger)

	// Se registra el controlador que ofrece el procedimiento remoto
	pb.RegisterAudioServiceServer(grpcServer, objControlador)

	fmt.Println("Servidor gRPC escuchando en :50051...")

	// Iniciar el servidor
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Error al iniciar servidor gRPC: %v", err)
	}
}
