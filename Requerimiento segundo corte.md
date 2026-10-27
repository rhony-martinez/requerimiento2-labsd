## Laboratorio de Sistemas distribuidos

## Requerimiento segundo corte

## Contexto del proyecto a realizar

En este Proyecto se reutilizará un cliente construido en javaScript, el cual se conecta con un servidor de Streaming para descargar vía WEB-GRPC fragmentos de audio correspondientes a un audio de tipo canción, ruido blanco, audio libro o podcast. Entre el cliente y el servidor de Streaming existirá un Servidor Envoy que traduce WEB-GRPC (utiliza http v 1.0 )a GRPC (utiliza http v 2.0).

## Opciones del cliente

- a) El cliente web se conectará con un microservicio de reacciones mediante webSockets. Luego de la conexión, los clientes pueden enviar reacciones asociadas al audio que está reproduciendo. El microservicio reenviará la reacción a los clientes que están reproduciendo el mismo audio mediante un callBack y utilizando la conexión previamente creada.

- b) El cliente puede ver los tipos de audios, seleccionar un tipo, ver la lista de audios, seleccionar un audio, ver sus detalles y finalmente reproducir mediante streaming el audio.

- c) El cliente puede seleccionar un plan de suscripción y pagar por ese plan. Considere utilizar tokens idempotentes y temporizadores para reenviar mensajes. Utilice ecos del lado del cliente y servidor para mostrar el token que está siendo utilizado, si el pago se realizó exitosamente y si se almaceno el pago.

- d) Un cliente puede ver los audios que ha reproducido, la cantidad de veces que los reprodujo.

- e) El cliente puede simular usar un token previamente utilizado.

- f) Un administrador web, puede registrar un nuevo audio enviado sus metadatos y el audio. Al momento de registrar un nuevo audio, se debe considerar que cada tipo tiene unos metadatos propios, por ejemplo

## Música

- Artista Principal: El intérprete o banda líder del proyecto.

- Álbum: El nombre del disco, EP o sencillo que agrupa la obra.

- Género Musical: La categoría rítmica o cultural (Pop, Rock, Jazz).

## Podcasts

- Nombre del Podcast: El título general del programa o serie.

- Título del Episodio: El nombre del tema específico de ese audio.

- Anfitrión (Host): La persona que dirige o presenta el contenido.

## Audiolibros

- Título del Libro: El nombre de la obra literaria completa.

- Autor: El escritor original del texto.

- Narrador: El actor de voz que realiza la lectura.

## Ruido Blanco

- Tipo de Sonido: La variante de frecuencia (Ruido Blanco, Marrón, Rosa).

- Fuente del Audio: El origen del sonido (Lluvia, Ventilador, Bosque).

- Uso Sugerido: La intención del audio (Dormir, Concentración, Meditación).

En el siguiente diagrama se muestran los clientes y servidores. El proxy envoy no debe ser desarrollado.

![Diagrama base](Diagrama.png)

Para realizar el proyecto se adjunta el servidor de streaming, el servidor envoy, la configuración del servidor envoy y los archivos JavaScript del cliente web. En el cliente web modifique los archivos index.html para crear la interface gráfica y el archivo funciones.js para registrar y des registrar el cliente de los canales del servidor de reacciones. Esta práctica se realizará en Linux o windows.

## Pre requisitos en linux

## 1) Instalar node.js

En el Proyecto necesitamos instalar node.js para permitir la gestión de dependencias con npm. Ejecutar los siguientes comandos.

```sudo apt update```

```sudo apt install nodejs npm```

## 2) Instalación de paquetes para usar GRPC en javaScript

```npm install grpc-web google-protobuf```

## google-protobuf,

Es un paquete que permite:

- Obtener las clases necesarias para serializar y deserializar mensajes Protocol Buffers en JavaScript

- Implementación oficial generada por Google

- Utilidades para trabajar con .proto compilados a .js

## grpc-web

Es la biblioteca oficial de Google para usar gRPC desde el navegador o desde aplicaciones web. Permite:

- Hacer llamadas gRPC sobre HTTP/1.1

- Usar un envoy para traducir del navegador al backend gRPC real

- Consumir servicios gRPC desde Angular, React, Vue, etc.

## 3) Utilizar un servidor envoy

El navegador NO soporta gRPC nativo, por estas razones:

- El navegador solo puede usar HTTP/1.1 en la mayoría de casos.

- gRPC usa HTTP/2 y streaming completo.

- Hay restricciones de CORS, headers y seguridad.

Por eso no es posible conectar Angular, React o JavaScript directamente a un servidor gRPC. Por tal movito, gRPC-Web es una versión adaptada para navegadores, pero necesita un proxy que traduzca las peticiones al backend. Envoy recibe gRPC-Web (desde el navegador, HTTP/1.1) y lo transforma en gRPC normal (HTTP/2) hacia el backend en Go. Para ejecutar el servidor envoy, primero debemos descargar el servidor de la plataforma y el archivo de configuración y posteriormente ejecutar el siguiente comando:

```./envoyServer -c envoyConfig2.yaml --disable-hot-restart```

## Imágenes de la aplicación funcionando

Para ejecutar la aplicación

Ejecute el servidor de Streaming con el siguiente comando:

```go run main/servidor.go```

Ejecute el servidor de envoy con el siguiente comando

```./envoyServer -c envoyConfig2.yaml --disable-hot-restart```

Lance el cliente

## Abra el archivo index.html en el navegador

Al abrir la página se muestra un input para ingresar el nombre de la canción y un select para establecer el formato de la canción

![al abrir la página](img1.png)

Cuando un usuario solicita un audio se envía una petición al servidor envoy que corre sobre el puerto 8080. Al recibir el primer fragmento el reproductor se vuelve visible.

![reproductor](img2.png)

Al pausar la canción se muestra un eco indicando que se ha pausado el audio. En el archivo funciones.js hay un listener que se ejecuta cada vez que el usuario pausa o reproduce una canción. Utilice este elemento para registrar o des registrar al usuario de los canales de notificaciones.

![ecos](img3.png)

En la consola del navegador se muestran ecos que establecen que se ha realizado una conexión con el servidor de streaming y se están recibiendo los fragmentos.

![ecos del navegador](img4.png)

En la consola del servidor envoy se muestran ecos que determinan que el servidor está arriba.

![ecos envoy](img5.png)

En la consola del servidor de Streaming se muestran ecos que determinan que el servidor está arriba.

![ecos servidor streaming](img6.png)

Si realiza un cambio en el cliente web (archivo cliente_receptor_fragmentos) debe ejecutar el comando npx webpack con el fin de descargar librerías y resolver imports, y posteriormente vuelva a abrir el archivo index.html

![cambios en cliente](img7.png)