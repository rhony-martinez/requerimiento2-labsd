package discovery

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"sync/atomic"

	consul "github.com/hashicorp/consul/api"
)

// Resolver consulta el directorio Consul para obtener la dirección de un servicio.
type Resolver struct {
	cliente  *consul.Client
	contador atomic.Uint64 // para el round-robin
}

func NewResolver(consulAddr string) (*Resolver, error) {
	cfg := consul.DefaultConfig()
	cfg.Address = consulAddr
	cliente, err := consul.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &Resolver{cliente: cliente}, nil
}

// Resolver devuelve la URL base de una instancia sana del servicio (round-robin).
// passingOnly=true descarta instancia cuyo health check está caído
func (r *Resolver) Resolver(nombreMS string) (*url.URL, error) {
	entradas, _, err := r.cliente.Health().Service(nombreMS, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("error consultando Consul: %w", err)
	}
	if len(entradas) == 0 {
		return nil, fmt.Errorf("no hay instancias disponibles de %q", nombreMS)
	}

	// Round-robin: cada llamada avanza una posición
	i := r.contador.Add(1) - 1
	e := entradas[i%uint64(len(entradas))]

	host := e.Service.Address
	if host == "" {
		host = e.Node.Address
	}
	return &url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, strconv.Itoa(e.Service.Port)),
	}, nil
}
