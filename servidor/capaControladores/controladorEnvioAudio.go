package capacontroladores

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc/peer"
	capafachadaservices "servidor.local/grpc-servidor/capaFachadaServices"
	"servidor.local/grpc-servidor/capalogger"
	pb "servidor.local/grpc-servidor/serviciosCancion"
)

type ControladorServidor struct {
	pb.UnimplementedAudioServiceServer
	logger *capalogger.Logger
}

func NewControladorServidor(logger *capalogger.Logger) *ControladorServidor {
	return &ControladorServidor{
		logger: logger,
	}
}

// Implementación del procedimiento remoto
func (thisC *ControladorServidor) EnviarCancionMedianteStream(
	req *pb.PeticionDTO, stream pb.AudioService_EnviarCancionMedianteStreamServer) error {

	//invocación a operación sincrona
	direcionCliente := ObtenerDireccionCliente(stream.Context())
	//invocación a operación asincrona
	go thisC.logger.AlmacenarSolicitud(req.Titulo, direcionCliente)
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
