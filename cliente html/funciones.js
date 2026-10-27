// funcion.js - wrapper simple para llamar a la implementación del streaming

function pedirCancion(titulo, formato) {
    // Prioridad de implementaciones conocidas
    if (typeof window.iniciar_streaming_cancion === 'function') {
        return window.iniciar_streaming_cancion(titulo, formato);
    }
    if (typeof window.iniciar_streaming_cancion_impl === 'function') {
        return window.iniciar_streaming_cancion_impl(titulo, formato);
    }
    if (typeof window.iniciarStreamGRPCImpl === 'function') {
        return window.iniciarStreamGRPCImpl(titulo, formato);
    }
    if (typeof window.iniciarStreamGRPC === 'function') {
        return window.iniciarStreamGRPC(titulo, formato);
    }

    console.error('No se encontró ninguna implementación de iniciar_streaming_cancion.');
    const d = document.getElementById('log');
    if (d) {
        const p = document.createElement('div');
        p.className = 'error';
        p.textContent = 'No se encontró ninguna implementación de iniciar_streaming_cancion.';
        d.appendChild(p);
    }
}

// Export para compatibilidad con cargas como módulo (si fuese necesario)
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { pedirCancion };
}

// --- Helpers y listeners para logging de audio (play / pause) ---
(function() {
    function writeLog(message, level) {
        const d = document.getElementById('log');
        if (!d) return;
        const p = document.createElement('div');
        p.className = level || '';
        const ts = new Date().toLocaleTimeString();
        p.textContent = `[${ts}] ${message}`;
        d.appendChild(p);
        d.scrollTop = d.scrollHeight;
    }

    function attachAudioListeners() {
        const audio = document.getElementById('audio-player');
        if (!audio) {
            writeLog('No se encontr\u00f3 el elemento audio#audio-player.', 'error');
            return;
        }
        audio.addEventListener('play', function() {
            writeLog('Reproducci\u00f3n iniciada (play).', 'success');
        });
        audio.addEventListener('pause', function() {
            writeLog('Reproducci\u00f3n pausada (pause).', 'error');
        });
    }

    // Intentar auto-adjuntar cuando el DOM est\u00e9 listo
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', attachAudioListeners);
    } else {
        attachAudioListeners();
    }

    // Exponer funciones para uso manual o tests
    if (typeof window !== 'undefined') {
        window.__writeAudioLog = writeLog;
        window.__attachAudioListeners = attachAudioListeners;
    }
})();
