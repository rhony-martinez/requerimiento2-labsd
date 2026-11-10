// Vista de administración: registrar un audio nuevo (multipart/form-data).

const MAX_MB = 50;

// Lee la duración del mp3 con un <audio> en memoria, para autocompletar el campo.
function leerDuracion(archivo) {
    return new Promise(resolver => {
        const url = URL.createObjectURL(archivo);
        const a = new Audio();
        a.preload = 'metadata';
        a.onloadedmetadata = () => {
            URL.revokeObjectURL(url);
            resolver(Number.isFinite(a.duration) ? Math.round(a.duration) : null);
        };
        a.onerror = () => { URL.revokeObjectURL(url); resolver(null); };
        a.src = url;
    });
}

async function mostrarAdmin() {
    try {
        const tipos = await apiGet('/api/catalogo/tipos');

        const clave = campo('Clave de administrador', sessionStorage.getItem('claveAdmin') || '');
        clave.input.type = 'password';

        const selector = document.createElement('select');
        tipos.forEach(t => {
            const o = document.createElement('option');
            o.value = t.id;
            o.textContent = t.nombre;
            selector.appendChild(o);
        });
        const filaTipo = el('div', '', 'campo');
        const etiquetaTipo = el('label', 'Tipo de audio');
        etiquetaTipo.appendChild(selector);
        filaTipo.appendChild(etiquetaTipo);

        const titulo = campo('Título', '');
        const duracion = campo('Duración (segundos)', '');
        duracion.input.type = 'number';
        duracion.input.min = '1';

        const entradaArchivo = document.createElement('input');
        entradaArchivo.type = 'file';
        entradaArchivo.accept = '.mp3,audio/mpeg';
        const filaArchivo = el('div', '', 'campo');
        const etiquetaArchivo = el('label', 'Archivo mp3');
        etiquetaArchivo.appendChild(entradaArchivo);
        filaArchivo.appendChild(etiquetaArchivo);

        const zonaCampos = el('div');      // metadatos propios del tipo
        let camposActuales = [];           // [{clave, input}]
        const mensaje = el('p', '', 'resultado-pago');

        function avisar(texto, esError) {
            mensaje.textContent = texto;
            mensaje.className = 'resultado-pago' + (esError ? ' error-pantalla' : '');
        }

        // Los campos los dicta el servidor: este cliente no conoce los metadatos de cada tipo.
        async function cargarCampos() {
            try {
                const defs = await apiGet(`/api/catalogo/tipos/${selector.value}/campos`);
                camposActuales = defs.map(d => ({ clave: d.clave, ...campo(d.etiqueta, '') }));
                zonaCampos.replaceChildren(...camposActuales.map(c => c.fila));
            } catch (e) { avisar(e.message, true); }
        }
        selector.addEventListener('change', cargarCampos);

        entradaArchivo.addEventListener('change', async () => {
            const f = entradaArchivo.files[0];
            if (!f) return;
            const seg = await leerDuracion(f);
            if (seg) duracion.input.value = seg;   // editable por si acaso
        });

        const btnRegistrar = boton('⬆ Registrar audio', registrar);

        async function registrar() {
            const archivo = entradaArchivo.files[0];
            if (!archivo) { avisar('Selecciona un archivo mp3.', true); return; }
            if (archivo.size > MAX_MB * 1024 * 1024) {
                avisar(`El archivo supera ${MAX_MB} MB.`, true);
                return;
            }
            const metadatos = {};
            camposActuales.forEach(c => { metadatos[c.clave] = c.input.value.trim(); });

            const form = new FormData();
            form.append('tipo_id', selector.value);
            form.append('titulo', titulo.input.value.trim());
            form.append('duracion_seg', duracion.input.value);
            form.append('metadatos', JSON.stringify(metadatos));
            form.append('archivo', archivo);       // el archivo al final

            btnRegistrar.disabled = true;
            avisar('Subiendo...', false);
            try {
                sessionStorage.setItem('claveAdmin', clave.input.value);
            } catch { /* opcional */ }

            try {
                const resp = await fetch(`${window.GATEWAY_URL}/api/catalogo/audios`, {
                    method: 'POST',
                    headers: { 'X-Admin-Token': clave.input.value },
                    // SIN Content-Type: el navegador lo pone con el "boundary" correcto.
                    body: form
                });
                const cuerpo = await resp.json().catch(() => ({}));
                if (!resp.ok) {
                    avisar(`❌ ${cuerpo.error || 'Error HTTP ' + resp.status}`, true);
                } else {
                    avisar(`✅ Audio registrado con id ${cuerpo.id}: ${cuerpo.titulo}`, false);
                    titulo.input.value = '';
                    duracion.input.value = '';
                    entradaArchivo.value = '';
                    camposActuales.forEach(c => { c.input.value = ''; });
                }
            } catch {
                avisar('❌ No se pudo contactar al API Gateway. ¿Está encendido?', true);
            } finally {
                btnRegistrar.disabled = false;
            }
        }

        mostrar(
            el('h2', 'Administración: registrar audio'),
            clave.fila, filaTipo, titulo.fila, zonaCampos, duracion.fila, filaArchivo,
            btnRegistrar, mensaje,
            boton('← Atrás', mostrarTipos)
        );
        cargarCampos();
    } catch (e) { mostrarError(e); }
}