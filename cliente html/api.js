// Capa de acceso a servicios REST (siempre a través del API Gateway)
async function apiGet(ruta) {
    let resp;
    try {
        resp = await fetch(`${window.GATEWAY_URL}${ruta}`);
    } catch (e) {
        throw new Error('No se pudo contactar al API Gateway. ¿Está encendido?');
    }
    if (!resp.ok) {
        // Los microservicios devuelven {"error": "..."}; lo mostramos tal cual
        const cuerpo = await resp.json().catch(() => ({}));
        throw new Error(cuerpo.error || `Error HTTP ${resp.status}`);
    }
    return resp.json();
}

// Identificación simple: un id numérico por navegador, editable por el usuario
function usuarioId() {
    let id = localStorage.getItem('usuarioId');
    if (!id) {
        id = String(Math.floor(1000 + Math.random() * 9000));
        localStorage.setItem('usuarioId', id);
    }
    return Number(id);
}

function iniciarSelectorUsuario() {
    const campo = document.getElementById('usuario-id');
    campo.value = usuarioId();
    campo.addEventListener('change', () => {
        const valor = parseInt(campo.value, 10);
        if (valor > 0) localStorage.setItem('usuarioId', String(valor));
        else campo.value = usuarioId(); // valor inválido: restauramos
    })
}