package config

// Ruta asocia un prefijo público del gateway con el nombre de un servicio en Consul.
type Ruta struct {
	Prefijo  string // ruta pública en el gateway, p. ej. /api/productos
	Servicio string // nombre con el que el microservicio se registró en Consul
	Destino  string // ruta interna en el microservicio, p. ej. /productos
}

var Rutas = []Ruta{
	{Prefijo: "/api/catalogo", Servicio: "ms-audios", Destino: "/catalogo"},
	{Prefijo: "/api/estadisticas", Servicio: "ms-estadisticas", Destino: "/estadisticas"},
	// !Ir agregando las demás rutas
}
