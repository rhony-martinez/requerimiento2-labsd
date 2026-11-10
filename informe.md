# Informe breve: segundo corte

Notas de diseño, decisiones y comparaciones. Es material de apoyo y no un entregable.

## 1. Resumen

El primer corte era una arquitectura de clientes y servidores en consola. En este segundo corte el cliente es web y todo pasa por un **API Gateway** que descubre los servicios en **Consul**. Hay cuatro microservicios REST/WebSocket en Go, un servidor de streaming gRPC detrás de Envoy y una cola RabbitMQ que desacopla las estadísticas.

## 2. Decisiones de diseño

| Decisión | Motivo |
|---|---|
| Gateway en el puerto 9000 | El 8080 ya es de Envoy |
| Streaming por gateway → Envoy | Fiel al diagrama. Exige `FlushInterval = -1` en el proxy, o los fragmentos se acumulan |
| CORS solo en el gateway | Con CORS en dos capas el navegador recibe la cabecera duplicada y rechaza la respuesta |
| Usuario y audio en metadatos gRPC | No se modifica el `.proto` ni se regeneran los stubs |
| Cola RabbitMQ para estadísticas | Si estadísticas cae, el streaming sigue y los mensajes esperan. Absorbe picos de reproducciones |
| Metadatos como `map[string]string` | Un solo modelo sirve para los cuatro tipos, y el servidor dicta los campos de cada uno |
| Archivo con nombre generado | Evita sobrescribir archivos y rutas como `../../x` |
| Ids de audio nunca reutilizados | El id une el catálogo con el historial. Reutilizarlo falsearía las estadísticas |
| Registro antes que archivo al borrar | El peor fallo posible es un mp3 huérfano, nunca un catálogo que apunta a un archivo inexistente |
| Duración calculada en el servidor | No se confía en el dato del administrador |
| JSON atómico (`.tmp` + `Rename`) | Un corte durante la escritura no corrompe el archivo |
| JSON Lines en el historial | Solo crece, así que se agrega al final sin reescribir todo |

## 3. Idempotencia en los pagos

- El token se pide **una vez** por pago y todos los reintentos lo reutilizan. Si cada reintento pidiera un token nuevo, un pago con la respuesta perdida se cobraría dos veces.
- Antes de cobrar, el servidor pregunta si ya existe una transacción con ese token. Si existe, no cobra y devuelve la original.
- Esa pregunta y el registro del pago están dentro de un `sync.Mutex`: dos peticiones simultáneas con el mismo token no pueden pagar dos veces.
- **Perder la petición no es perder la respuesta.** En el primer caso no hubo cobro. En el segundo sí, y el cliente no se entera. Por eso existe la consulta de transacciones.
- Reintentos con espera exponencial (2, 4, 8 y 16 s) y timeout de 3 s por intento. Se reintentan los `5xx` y los errores de red, pero no los `4xx`.
- Pulsar «Pagar» dos veces **sí** genera dos tokens y dos cobros. La idempotencia protege los reintentos de un mismo pago, no a un usuario que inicia dos. Por eso el botón se bloquea durante el pago.
- Limitación: el mutex vive en un proceso. Con varias réplicas haría falta una restricción única en una base de datos compartida.

## 4. Callback por WebSocket frente a Java RMI

Es la comparación que pide la guía de reacciones. Aquí se resuelve con un hub en memoria y salas por audio. En RMI, el servidor invoca un método de un objeto remoto exportado por el cliente.

| Aspecto | Java RMI | WebSocket (este proyecto) |
|---|---|---|
| **Quién abre la conexión** | El servidor debe poder alcanzar al cliente (el cliente exporta un objeto remoto) | El cliente abre la conexión y el servidor responde por ella |
| **Firewalls y NAT** | Problemáticos: el servidor necesita llegar al cliente y usa puertos dinámicos | Sin problema: un `GET` HTTP que se *mejora* (upgrade) y usa los puertos 80/443 |
| **Navegador** | No se puede usar | Soporte nativo |
| **Lenguaje** | Atado a Java | Cualquier lenguaje. El contrato es JSON |
| **Acoplamiento** | Interfaces remotas compartidas (`Remote`). Un cambio obliga a recompilar | Mensajes JSON con un campo `tipo`. Es más laxo, pero el contrato no lo verifica el compilador |
| **Registro del cliente** | Hay que registrar y des-registrar el callback a mano | Conectarse = suscribirse. El cierre se detecta solo, con `onclose` y ping/pong |
| **Detectar clientes muertos** | Excepciones al invocar (`RemoteException`) | Ping/pong periódico y *read deadline* |
| **Modelo de comunicación** | Llamada a procedimiento (síncrona, con valor de retorno) | Mensajes asíncronos en ambos sentidos (full-duplex) |
| **Agrupación (salas)** | Hay que construirla: lista de callbacks por audio | Igual: mapa `audio → clientes`. En STOMP/Spring vendría con el broker |
| **Cliente lento** | Una llamada lenta puede bloquear al servidor | Un buzón por cliente. Si se llena se descarta, y los demás no se ven afectados |
| **Reconexión** | Manual | Simple: reintento con espera creciente y volver a la sala |

**Qué se mantiene igual:** en ambos el servidor guarda la lista de clientes interesados y los recorre para notificarlos, y en ambos la desconexión abrupta es el caso difícil.

**Conclusión:** para una web, WebSocket es la opción natural. RMI sirve entre aplicaciones Java en una red controlada.

Detalles de la implementación que valen la pena:
- Una conexión por navegador. Unirse o salir de una sala son mensajes, no conexiones nuevas.
- El hub no conoce WebSocket (solo clientes con un buzón), así que se puede probar solo.
- Un lector y un escritor por conexión, porque gorilla no admite escrituras concurrentes.
- Lista blanca de reacciones y límite de 5 por segundo por conexión (token bucket).
- El conteo de oyentes se encola dentro del lock para que llegue en el mismo orden en que cambió.

## 5. Cola de mensajes: lo que se gana y lo que se pierde

| | REST directo | Cola |
|---|---|---|
| Acoplamiento | Estadísticas debe estar vivo | Independientes |
| Picos | El destino se satura | Se encolan y se consumen a su ritmo |
| Consistencia | Inmediata | Eventual: el historial tarda un instante |
| Entrega | Falla o no falla | Al menos una vez: puede haber duplicados |
| Complejidad | Baja | Otro componente que operar |

El consumidor confirma (`Ack`) después de guardar. Los mensajes ilegibles o inválidos se descartan para que no se reintenten indefinidamente. Si falla el disco, se reencolan.

## 6. Problemas que aparecieron

- **`audio=0` en los metadatos:** un nombre mal escrito (`audioActualId` en vez de `audioActual`). Se halló comparando lo que el cliente enviaba, lo que recibía el servidor y lo que había en las cabeceras.
- **`Unlock of unlocked RWMutex`:** mezclar `RLock` con `Unlock`. Es un `fatal error` que Gin no atrapa.
- **504 de Envoy:** una petición gRPC sin cuerpo espera hasta el timeout de 15 s. Para streams largos conviene `timeout: 0s` en la ruta.
- **Conexión AMQP muerta tras recrear el contenedor:** el publicador conectaba una sola vez. Ahora conecta de forma perezosa y reconecta. El consumidor reintenta en bucle.
- **Salir de la sala solo con pausa:** `reacciones.js` no exponía `salir` en `window`, y `detener_streaming` (dentro del bundle) no podía llamarla.

## 7. Pendiente y mejoras

- Streaming real con Media Source Extensions en vez de acumular todo el audio.
- Respaldo local (outbox) para no perder reproducciones si RabbitMQ cae, e id único por mensaje para descartar duplicados.
- Idempotencia en el registro de audios con `Idempotency-Key`.
- Evento de «audio eliminado» hacia estadísticas.
- Pruebas con `go test -race` (se omitieron por no instalar cgo en Windows).
- TLS y una autenticación de administrador real.
