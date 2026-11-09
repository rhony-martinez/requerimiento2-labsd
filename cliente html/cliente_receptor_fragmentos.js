// cliente_receptor_fragmentos.js

const { peticionDTO } = require('./stubs_generados/servicios_pb.js');
const { AudioServicePromiseClient } = require('./stubs_generados/servicios_grpc_web_pb.js');



const GATEWAY_URL = window.GATEWAY_URL ||  'http://localhost:9000';

const client = new AudioServicePromiseClient(GATEWAY_URL);



// 1. Declara un array para almacenar los fragmentos de audio

// let audioChunks = [];

// Evita que los audios se mezclen cuando un usuario elige otro mientras aún no llega este
let streamActual = null;



// 2. Obtén la referencia al elemento <audio>

const audioPlayer = document.getElementById('audio-player');
let _firstFragmentReceived = false;



function iniciar_streaming_cancion(titulo, formato) {

    console.log("Iniciando llamada gRPC-Web a través de Envoy...");

    if (streamActual) { streamActual.cancel(); }
    audioChunks = [];
    _firstFragmentReceived = false;
    if (audioPlayer) { audioPlayer.pause(); }

    const request = new peticionDTO();

    request.setTitulo(titulo);    

    request.setFormato(formato);

    console.log("Solicitud preparada:", request.toObject());
    const metadatos = {
        'x-usuario-id': String(window.usuarioId ? window.usuarioId() : 0),
        'x-audio-id': String(window.audioActual ? window.audioActual.id : 0)
    }
    const stream = client.enviarCancionMedianteStream(request, metadatos)
    streamActual = stream;



    // Maneja la llegada de cada fragmento

    stream.on('data', (response) => {
        if (stream !== streamActual) return; // Ignorar eventos de un stream cancelado

        // Obtiene el array de bytes (Uint8Array)

        const data = response.getData_asU8(); 

        console.log(`Fragmento recibido desde el servidor. Tamaño: ${data.length} bytes`);

        // Almacena el fragmento en el array
        audioChunks.push(data);

        // Si es el primer fragmento, mostramos el reproductor y escribimos en el log
        if (!_firstFragmentReceived) {
            _firstFragmentReceived = true;
            if (audioPlayer) {
                audioPlayer.style.display = '';
            }
            if (typeof window !== 'undefined' && window.__writeAudioLog) {
                window.__writeAudioLog('Primer fragmento recibido. Reproductor mostrado.', 'success');
            } else {
                console.log('Primer fragmento recibido. Reproductor mostrado.');
            }
        }

    });



    // Maneja el final del stream

    stream.on('end', () => {
        if (stream !== streamActual) return; // Ignorar eventos de un stream cancelado

        console.log('✅ Transmisión de la canción finalizada. Preparando reproducción...');

        // Combina todos los fragmentos en un solo Blob
        const audioBlob = new Blob(audioChunks, { type: 'audio/mpeg' }); 

        // Crea una URL de objeto para el Blob
        const audioUrl = URL.createObjectURL(audioBlob);

        // Asigna la URL al reproductor y comienza la reproducción
        if (audioPlayer) {
            audioPlayer.src = audioUrl;
            audioPlayer.play()
                .then(() => {
                    console.log('▶️ Reproducción de la canción iniciada con éxito.');
                })
                .catch(error => {
                    console.error('❌ Error al iniciar la reproducción automática. Intente dar clic en el reproductor.', error);
                });
        }

        // Limpia los fragmentos para una próxima llamada
        audioChunks = [];
        // reset flag for future requests
        _firstFragmentReceived = false;
    });

    stream.on('error', (err) => {
        if (stream !== streamActual) return; // Ignorar eventos de un stream cancelado

        console.error('ERROR en la comunicación gRPC-Web:', err);

    });

}

function detener_streaming() {
    // 1. Cancelar la descarga en curso, si la hay
    if (streamActual) {
        streamActual.cancel();
        // Al poner null, los callbacks que aún lleguen se ignoran
        // gracias a la guarda "if (stream !== streamActual) return;"
        streamActual = null;
    }

    // 2. Descartar lo acumulado
    audioChunks = [];
    _firstFragmentReceived = false;

    // 3. Detener y ocultar el reproductor
    if (audioPlayer) {
        audioPlayer.pause();
        audioPlayer.removeAttribute('src'); // suelta el audio actual
        audioPlayer.load();                 // restablece el elemento
        audioPlayer.style.display = 'none';
    }

    const estado = document.getElementById('status');
    if (estado) estado.textContent = 'Reproducción detenida';
}

// Exponerla para que catalogo.js (que está fuera del bundle) pueda llamarla
if (typeof window !== 'undefined') {
    window.detener_streaming = detener_streaming;
}

// Export implementation to global window for consumers (index.html / funcion.js)

if (typeof window !== 'undefined') {

    window.iniciar_streaming_cancion = iniciar_streaming_cancion;

    // backward compatible name

    window.iniciarStreamGRPCImpl = iniciar_streaming_cancion;

    console.log('Exported iniciar_streaming_cancion to window');

}
