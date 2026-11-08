package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

// HandlerEnvoy reenvía el tráfico de gRPC-Web hacia Envoy
// A diferencia de Handler (que resuelve en Consul por petición), aquí el destino es fijo: Envoy es infraestructura suministrada, no un microservicio nuestro
// Por eso, el ReverseProxy se crea una sola vez
func HandlerEnvoy(envoyURL *url.URL) gin.HandlerFunc {
	rp := &httputil.ReverseProxy{
		// Imprescindible para streaming: envía cada fragmento apenas llega
		FlushInterval: -1,
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(envoyURL)
			// Envoy tiene su propia política CORS. Si le llega el header Origin,
			// agrega Access-Control-Allow-Origin, y el gateway ya agrega el suyo:
			// el navegador vería la cabecera duplicada y rechazaría la respuesta
			// Quitando Origin, Envoy no hace CORS y el gateway queda como único responsable
			pr.Out.Header.Del("Origin")
			pr.SetXForwarded()
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("[gateway] error hacia Envoy: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`{"error":"el proxy Envoy no respondió"}`))
		},
	}

	return func(c *gin.Context) {
		log.Printf("[gateway] gRPC-Web %s -> Envoy (%s)", c.Request.URL.Path, envoyURL)
		rp.ServeHTTP(c.Writer, c.Request)
	}
}
