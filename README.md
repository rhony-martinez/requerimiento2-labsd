# Spotify distribuido: segundo corte

Aplicación web de audio (música, podcasts, audiolibros y ruido blanco) construida como un conjunto de microservicios en Go. El navegador solo habla con un **API Gateway**; este descubre los servicios en **Consul** y reenvía las peticiones.

## Qué hace

| Opción | Funcionalidad | Servicio |
|---|---|---|
| a | Reacciones en tiempo real (WebSocket), solo para quienes escuchan el mismo audio | `ms-reacciones` |
| b | Catálogo: tipos, lista, detalle y reproducción por streaming | `ms-audios` + streaming |
| c | Suscripciones y pago con token idempotente, con reintentos | `ms-suscripciones` |
| d | Historial de reproducciones con conteo | `ms-estadisticas` |
| e | Reenviar un pago con un token ya utilizado | `ms-suscripciones` + cliente |
| f | Administrador: registrar y eliminar audios con sus metadatos | `ms-audios` |

## Arquitectura

```
Cliente web (:3000) ──► API Gateway (:9000) ──► Consul (:8500)  (descubrimiento)
                              │
                              ├─► ms-audios         (:8081)  REST: catálogo y administración
                              ├─► ms-reacciones     (:8082)  WebSocket
                              ├─► ms-suscripciones  (:8083)  REST: pagos idempotentes
                              ├─► ms-estadisticas   (:8084)  REST: historial
                              └─► Envoy (:8080) ──► Servidor de streaming (:50051, gRPC)
                                                           │ publica (asíncrono)
                                                           ▼
                                                  RabbitMQ (:5672) ──► ms-estadisticas
```

Puntos clave:
- **Seguridad:** el gateway agrega la cabecera `X-Gateway-Token`; cada microservicio rechaza con 403 lo que no la traiga. Las operaciones de administrador exigen además `X-Admin-Token`.
- **CORS:** lo pone únicamente el gateway.
- **Streaming:** gRPC-Web desde el navegador, a través del gateway y Envoy, hasta el servidor Go. El usuario y el audio viajan como cabeceras (metadatos gRPC).
- **Estadísticas:** el streaming publica cada reproducción en una cola y no espera respuesta. Si `ms-estadisticas` está caído, los mensajes esperan en RabbitMQ.
- **Capas:** controladores → fachada (services) → repositorios, con DTO para lo que viaja por la red.
- **Ecos:** toda llamada REST, gRPC, WebSocket o consumo de cola imprime una línea en la consola del servidor (`[REST]`, `[gRPC]`, `[WS]`, `[COLA]`, `[TOKEN]`, `[PAGO]`).

## Requisitos

- Go 1.23 o superior
- Node.js LTS (incluye npm)
- Docker Desktop (Consul, Envoy y RabbitMQ)
- `curl.exe` (incluido en Windows 11)

## Estructura

```
├── servidor/            Streaming gRPC (puerto 50051) y carpeta canciones/
├── cliente html/        Cliente web (HTML, JS, webpack)
├── api-gateway/         Gateway (9000)
├── ms-audios/           Catálogo y administración (8081)
├── ms-reacciones/       Reacciones por WebSocket (8082)
├── ms-suscripciones/    Pasarela de pagos idempotente (8083)
├── ms-estadisticas/     Consumidor de la cola e historial (8084)
├── docker-compose.yml   Consul, Envoy y RabbitMQ
└── envoyConfig2.yaml    Configuración de Envoy (suministrada)
```

## Cómo ejecutarlo

Abre una terminal por servicio y respeta este orden.

**1. Infraestructura**
```powershell
docker compose up -d
```
Consul: http://localhost:8500 · RabbitMQ: http://localhost:15672 (guest / guest)

**2. Microservicios** (cada uno en su carpeta)
```powershell
cd ms-audios;         go run .
cd ms-estadisticas;   go run .
cd ms-reacciones;     go run .
cd ms-suscripciones;  go run .
```

**3. Servidor de streaming**
```powershell
cd servidor
go run main/servidor.go
```

**4. API Gateway**
```powershell
cd api-gateway
go run .
```

**5. Cliente web**
```powershell
cd "cliente html"
npm install
npx webpack          # solo la primera vez o si cambia cliente_receptor_fragmentos.js
npx serve -l 3000 .
```
Abre http://localhost:3000. Tras cambiar un `.js` que no pasa por webpack, basta Ctrl+F5.

En Consul (pestaña *Services*) deben aparecer `ms-audios`, `ms-estadisticas`, `ms-reacciones` y `ms-suscripciones`.

## Configuración

Todo tiene valor por defecto. Solo cambia lo siguiente si lo necesitas.

| Variable | Defecto | Uso |
|---|---|---|
| `PORT` | según servicio | Puerto HTTP |
| `CONSUL_ADDR` | `127.0.0.1:8500` | Dirección de Consul |
| `GATEWAY_SECRET` | `secreto-gateway-2026` | Debe ser igual en gateway y microservicios |
| `ADMIN_TOKEN` | `admin-2026` | Clave de administrador (`ms-audios`) |
| `ENVOY_ADDR` | `http://127.0.0.1:8080` | Destino gRPC-Web del gateway |
| `AMQP_URL` | `amqp://guest:guest@127.0.0.1:5672/` | RabbitMQ (`ms-estadisticas`) |
| `DIRECTORIO_AUDIOS` | `../servidor/canciones` | Dónde guarda los mp3 `ms-audios` |

## Uso del cliente

- **Tipos de audio:** elige un tipo, un audio y pulsa ▶ Reproducir. Con el audio sonando se activan las reacciones y se muestra cuántos oyentes hay.
- **📜 Mi historial:** audios reproducidos y cuántas veces.
- **💳 Suscripciones:** elige un plan y paga. La bitácora muestra el token y los reintentos. «♻ Reenviar pago con el último token» simula reutilizar un token (opción e). «🧾 Mis pagos» consulta las transacciones por cédula.
- **🛠 Administración** (clave `admin-2026`): registra un audio con su mp3 y los metadatos propios del tipo (los campos los dicta el servidor, y la duración se calcula en el servidor). «🗑 Gestionar audios» los elimina.

El id de usuario (arriba a la derecha) identifica el historial y las reacciones. Para probar reacciones abre dos pestañas con el mismo audio.

## Probar la pasarela de pagos

Los casos de fallo se activan con variables de entorno al lanzar `ms-suscripciones`:

```powershell
$env:MODO_FALLO="NINGUNO"; go run .                                  # 1 y 5: éxito / token ya utilizado
$env:MODO_FALLO="PETICION"; $env:CANTIDAD_FALLOS="3"; go run .       # 2: éxito en el reintento 3
$env:MODO_FALLO="PETICION"; $env:CANTIDAD_FALLOS="5"; go run .       # 3: se pierden las peticiones, sin cobro
$env:MODO_FALLO="RESPUESTA"; $env:CANTIDAD_FALLOS="5"; go run .      # 4: se pierden las respuestas, sí hay cobro
```
Ejecútalos en orden. Para repetir desde cero, detén el servicio y borra `ms-suscripciones/datos/`.

## Pruebas rápidas con curl

```powershell
curl.exe http://localhost:9000/api/catalogo/tipos                    # 200
curl.exe http://localhost:8081/catalogo/tipos                        # 403: directo, sin pasar por el gateway
curl.exe -i -X DELETE http://localhost:9000/api/catalogo/audios/1    # 401: falta la clave de administrador
curl.exe http://localhost:9000/api/estadisticas/usuarios/1/historial # 200
```
No agregues `X-Admin-Token` a la prueba del `DELETE`, o borrarías un audio real.

## Datos

- Catálogo: `ms-audios/datos/audios.json` (se crea en el primer arranque con los 4 tipos).
- Historial: `ms-estadisticas/datos/reproducciones.jsonl`.
- Pagos: `ms-suscripciones/datos/`.
- Audios: `servidor/canciones/`.

El catálogo y el historial comparten ids de audio, así que si reinicias uno, reinicia el otro. Con los servicios detenidos, borra las carpetas `datos/` correspondientes.

## Limitaciones conocidas

- El cliente descarga todo el audio antes de reproducirlo. El streaming real (Media Source Extensions) queda como mejora.
- Si RabbitMQ está caído al reproducir, esa reproducción se pierde (no hay respaldo local).
- La cola entrega al menos una vez, por lo que un mensaje duplicado sumaría una reproducción de más.
- El mutex de la pasarela solo protege una instancia. Con varias réplicas haría falta una restricción única en una base de datos compartida.
- Envoy sigue accesible directamente en `127.0.0.1:8080`.
- Todo viaja por HTTP plano. En un sistema real se usaría TLS.
