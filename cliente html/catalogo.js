// Vistas del catálogo: Tipos -> Lista de audios -> Detalle -> Reproducir

const pantalla = document.getElementById('pantalla');

// -------------- helpers de DOM ----------------
function el(etiqueta, texto, clase) {
    const nodo = document.createElement(etiqueta);
    if (texto) nodo.textContent = texto;
    if (clase) nodo.className = clase;
    return nodo;
}

function boton(texto, alClic) {
    const b = el('button', texto);
    b.addEventListener('click', alClic);
    return b;
}

function mostrar(...nodos) {
    pantalla.replaceChildren(...nodos);
}

function formatearDuracion(seg) {
    const m = Math.floor(seg / 60);
    const s = String(seg % 60).padStart(2, '0');
    return `${m}:${s}`;
}

// "artista_principal" -> "Artista principal"
function etiquetaMetadato(clave) {
    const t = clave.replaceAll('_', ' ')
    return t.charAt(0).toUpperCase() + t.slice(1);
}

function mostrarError(e) {
    console.error(e); // la traza real, con archivo y línea
    mostrar(el('p', e.message, 'error-pantalla'), boton('Reintentar', mostrarTipos));
}

// -------------- vistas ----------------
async function mostrarTipos() {
    try {
        const tipos = await apiGet('/api/catalogo/tipos');
        const lista = el('ul', '', 'menu');
        tipos.forEach(t => {
            const li = el('li');
            li.appendChild(boton(t.nombre, () => mostrarAudios(t)));
            lista.appendChild(li);
        });
        mostrar(el('h2', 'Tipos de audio'), lista, boton('📜 Mi historial', mostrarHistorial));
    } catch (e) { mostrarError(e); }
}

async function mostrarAudios(tipo) {
    try {
        const audios = await apiGet(`/api/catalogo/tipos/${tipo.id}/audios`);
        const lista = el('ul', '', 'menu');
        audios.forEach(a => {
            const li = el('li');
            li.appendChild(boton(a.titulo, () => mostrarDetalle(a.id, tipo)));
            lista.appendChild(li);
        });
        mostrar(el('h2', `Tipo: ${tipo.nombre}`), lista, boton('← Atrás', mostrarTipos));
    } catch (e) { mostrarError (e); }
}

async function mostrarDetalle(audioId, tipo) {
    try {
        const d = await apiGet(`/api/catalogo/audios/${audioId}`);

        const dl = el('dl', '', 'detalle');
        const agregar = (etiqueta, valor) => {
            dl.appendChild(el('dt', etiqueta));
            dl.appendChild(el('dd', valor));
        };
        agregar('Título', d.titulo);
        agregar('Tipo', d.tipo);
        agregar('Duración', formatearDuracion(d.duracion_seg));
        Object.entries(d.metadatos || {}).forEach(([k, v]) => agregar(etiquetaMetadato(k), v));

        mostrar(
            el('h2', `${d.tipo}: ${d.titulo}`),
            dl,
            boton('▶ Reproducir', () => reproducir(d)),
            boton('■ Detener', () => detener_streaming()),
            boton('← Atrás', () => { detener_streaming(); mostrarAudios(tipo); })
        )
    } catch (e) { mostrarError(e); }
}

async function mostrarHistorial() {
    try {
        const items = await apiGet(`/api/estadisticas/usuarios/${usuarioId()}/historial`);
        const atras = boton('← Atrás', mostrarTipos);

        if (items.length === 0) {
            mostrar(el('h2', 'Mi historial'), el('p', 'Aún no has reproducido ningún audio.'), atras);
            return;
        }

        // El servicio de estadísticas solo conoce ids; los títulos los pide el cliente al catálogo
        const detalles = await Promise.all(
            items.map(i => apiGet(`/api/catalogo/audios/${i.audio_id}`).catch(() => null))
        );

        const tabla = el('table');
        const cabecera = el('tr');
        ['Audio', 'Tipo', 'Veces', 'Última vez'].forEach(t => cabecera.appendChild(el('th', t)));
        tabla.appendChild(cabecera);

        items.forEach((item, i) => {
            const d = detalles[i]
            const fila = el('tr');
            fila.appendChild(el('td', d ? d.titulo : `Audio #${item.audio_id}`));
            fila.appendChild(el('td', d ? d.tipo : '-'));
            fila.appendChild(el('td', String(item.reproducciones)));
            fila.appendChild(el('td', new Date(item.ultima_fecha).toLocaleString()));
            tabla.appendChild(fila);
        })

        mostrar(el('h2', 'Mi historial'), tabla, atras);
    } catch (e) { mostrarError(e); }
}

function reproducir(detalle) {
    console.log('Reproduciendo:', detalle); 
    // La Fase 5 (reacciones) y la 4 (estadísticas) necesitas saber qué audio suena
    window.audioActual = { id: detalle.id, titulo: detalle.titulo };
    document.getElementById('status').textContent = `Cargando: ${detalle.titulo}...`;
    // El streaming busca por nombre de archivo, por eso enviamos "archivo", no "titulo"
    pedirCancion(detalle.archivo, detalle.formato)
}

// ------------------ arranque -------------
iniciarSelectorUsuario();
mostrarTipos();