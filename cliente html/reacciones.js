// Cliente de reacciones: una conexión WebSocket, y la pertenencia a la sala
// de un audio depende de si el <audio> se está reproduciendo.

const etiquetaOyentes = document.getElementById('oyentes');

(function () {
    const REACCIONES = { like: '👍', heart: '❤️', fun: '😂', clap: '👏', fire: '🔥' };

    let socket = null;
    let audioEnSala = null;     // el audio en que DEBERÍAMOS estar (sirve para re-unirse)
    let esperaMs = 1000;        // backoff de reconexión

    const barra = document.getElementById('barra-reacciones');
    const contenedor = document.getElementById('contenedor-reacciones');
    const audio = document.getElementById('audio-player');

    function urlWS() {
        // http -> ws, https -> wss
        const base = window.GATEWAY_URL.replace(/^http/, 'ws');
        return `${base}/api/reacciones/ws?usuario_id=${usuarioId()}`;
    }

    function log(msg, nivel) {
        if (window.__writeAudioLog) window.__writeAudioLog(msg, nivel);
    }

    function conectar() {
        socket = new WebSocket(urlWS());

        socket.onopen = () => {
            esperaMs = 1000;
            console.log('[reacciones] conectado');
            // Si se cayó la conexión mientras sonaba algo, volvemos a la sala.
            if (audioEnSala !== null) enviar({ tipo: 'unirse', audio_id: audioEnSala });
        };

        socket.onmessage = (evento) => {
            let m;
            try { m = JSON.parse(evento.data); } catch { return; }
            if (m.tipo === 'reaccion' && m.audio_id === audioEnSala) {
                animar(m.reaccion);
                log(`Usuario ${m.usuario_id} reaccionó ${REACCIONES[m.reaccion] || ''}`, 'success');
            } if (m.tipo === 'error') {
                console.warn('[reacciones] error del servidor:', m.error);
            } else if (m.tipo === 'oyentes' && m.audio_id === audioEnSala) {
                mostrarOyentes(m.cantidad);
            }
        };

        socket.onclose = () => {
            console.log(`[reacciones] desconectado; reintento en ${esperaMs} ms`);
            setTimeout(conectar, esperaMs);
            esperaMs = Math.min(esperaMs * 2, 10000);
        };
        // onerror siempre va seguido de onclose, que es quien reconecta.
    }

    function enviar(objeto) {
        if (socket && socket.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify(objeto));
        }
    }

    function unirse(audioId) {
        audioEnSala = audioId;
        enviar({ tipo: 'unirse', audio_id: audioId });
        habilitarBotones(true);
    }

    function salir() {
        console.log('salir()')
        if (audioEnSala === null) return;
        audioEnSala = null;
        enviar({ tipo: 'salir' });
        mostrarOyentes(null);
        habilitarBotones(false);
    }

    function animar(clave) {
        const burbuja = document.createElement('div');
        burbuja.className = 'reaccion-burbuja';
        burbuja.textContent = REACCIONES[clave] || '❓';
        burbuja.style.left = `${Math.random() * 90}%`;
        contenedor.appendChild(burbuja);
        setTimeout(() => burbuja.remove(), 3000);
    }

    function reaccionar(clave) {
        if (audioEnSala === null) return;
        animar(clave);                         // animación local inmediata
        enviar({ tipo: 'reaccion', reaccion: clave });
    }

    function habilitarBotones(si) {
        barra.querySelectorAll('button').forEach(b => { b.disabled = !si; });
    }

    // ---- arranque ----
    Object.entries(REACCIONES).forEach(([clave, emoji]) => {
        const b = document.createElement('button');
        b.textContent = emoji;
        b.title = clave;
        b.disabled = true;
        b.addEventListener('click', () => reaccionar(clave));
        barra.appendChild(b);
    });

    // Reproduciendo = escuchando = en la sala. Salir es idempotente,
    // así que pause y ended pueden dispararse juntos sin problema.
    audio.addEventListener('play', () => {
        if (window.audioActual) unirse(window.audioActual.id);
    });
    audio.addEventListener('pause', salir);
    audio.addEventListener('ended', salir);

    // Si el usuario cambia su id, reconectamos para que el servidor vea el nuevo.
    document.getElementById('usuario-id').addEventListener('change', () => {
        if (socket) socket.close(); // onclose reconecta con el id nuevo
    });

    conectar();
})();

function mostrarOyentes(n) {
    etiquetaOyentes.textContent = (n === null)
        ? ''
        : `👥 ${n} oyente${n === 1 ? '' : 's'} en este audio`;
}