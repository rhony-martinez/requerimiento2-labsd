package capacontroladores

import (
	"fmt"
	"strconv"

	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	capafachadaservices "servidor.local/grpc-servidor/capaFachadaServices"
	"servidor.local/grpc-servidor/capalogger"
	"servidor.local/grpc-servidor/capamensajeria"
	pb "servidor.local/grpc-servidor/serviciosCancion"
)

type ControladorServidor struct {
	pb.UnimplementedAudioServiceServer
	logger     *capalogger.Logger
	publicador *capamensajeria.Publicador
}

func NewControladorServidor(logger *capalogger.Logger, pub *capamensajeria.Publicador) *ControladorServidor {
	return &ControladorServidor{
		logger:     logger,
		publicador: pub,
	}
}

// Implementación del procedimiento remoto
func (thisC *ControladorServidor) EnviarCancionMedianteStream(
	req *pb.PeticionDTO, stream pb.AudioService_EnviarCancionMedianteStreamServer) error {

	//invocación a operación sincrona
	direcionCliente := ObtenerDireccionCliente(stream.Context())
	//invocación a operación asincrona
	go thisC.logger.AlmacenarSolicitud(req.Titulo, direcionCliente)

	// Vemos quién reproduce qué de los metadatos gRPC (claves en minúscula)
	usuarioID, audioID := leerIdentificadores(stream.Context())
	fmt.Printf("[gRPC] enviarCancionMedianteStream archivo=%s usuario=%d audio=%d\n", req.GetTitulo(), usuarioID, audioID)
	if usuarioID > 0 && audioID > 0 {
		go thisC.publicador.PublicarReproduccion(usuarioID, audioID) // asíncrono
	}

	combined := req.GetTitulo() + "." + req.GetFormato()
	return capafachadaservices.StreamAudioFile(
		combined,
		func(data []byte) error {
			return stream.Send(&pb.FragmentoCancion{Data: data})
		})

}

func ObtenerDireccionCliente(ctx context.Context) string {
	if p, ok := peer.FromContext(ctx); ok {
		return p.Addr.String()
	}
	return "desconocido"
}

func leerIdentificadores(ctx context.Context) (usuarioID, audioID int) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, 0
	}
	leer := func(clave string) int {
		valores := md.Get(clave)
		if len(valores) == 0 {
			return 0
		}
		n, _ := strconv.Atoi(valores[0]) // si no es número queda en 0
		return n
	}
	return leer("x-usuario-id"), leer("x-audio-id")
}
