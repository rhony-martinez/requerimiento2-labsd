// pagos.js: suscripciones y pagos idempotentes

const PAGO_TIMEOUT_MS = 3000;
const ESPERAS_MS = [2000, 4000, 8000, 16000];   // backoff exponencial
const CLAVE_ULTIMO_TOKEN = 'ultimoToken';

function esperar(ms) {
    return new Promise(resolver => setTimeout(resolver, ms));
}

// POST JSON con tiempo máximo de espera. Lanza error si no hay respuesta.
async function postJSON(ruta, cuerpo, timeoutMs) {
    const control = new AbortController();
    const temporizador = setTimeout(() => control.abort(), timeoutMs);
    try {
        const resp = await fetch(`${window.GATEWAY_URL}${ruta}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(cuerpo),
            signal: control.signal
        });
        const datos = await resp.json().catch(() => ({}));
        return { estado: resp.status, datos };
    } finally {
        clearTimeout(temporizador);
    }
}

function guardarUltimoToken(t) {
    try { localStorage.setItem(CLAVE_ULTIMO_TOKEN, JSON.stringify(t)); } catch { /* opcional */ }
}
function leerUltimoToken() {
    try { return JSON.parse(localStorage.getItem(CLAVE_ULTIMO_TOKEN)); } catch { return null; }
}

// Envía el pago: intento inicial + hasta 4 reintentos, SIEMPRE con el mismo token.
// Devuelve {estado, datos, intento} si el servidor respondió, o null si se agotaron.
async function enviarPagoConReintentos(token, monto, escribir) {
    const inicio = Date.now();
    const intentos = ESPERAS_MS.length + 1;

    for (let intento = 0; intento < intentos; intento++) {
        const nombre = intento === 0 ? 'Intento inicial' : `Reintento ${intento}`;
        const seg = ((Date.now() - inicio) / 1000).toFixed(1);
        escribir(`${nombre} (t=${seg}s): enviando pago con token ${token}`, '');

        try {
            const r = await postJSON('/api/suscripciones/pagos',
                { token_idempotencia: token, monto }, PAGO_TIMEOUT_MS);
            if (r.estado < 500) {
                return { ...r, intento };         // respuesta definitiva
            }
            escribir(`   el servidor respondió ${r.estado}`, 'error');
        } catch (e) {
            const motivo = e.name === 'AbortError' ? 'tiempo de espera agotado (3 s)' : 'error de red';
            escribir(`   sin respuesta: ${motivo}`, 'error');
        }

        if (intento < ESPERAS_MS.length) {
            escribir(`   esperando ${ESPERAS_MS[intento] / 1000} s antes de reintentar...`, '');
            await esperar(ESPERAS_MS[intento]);
        }
    }
    return null;
}

// ---------- vistas ----------

function crearBitacora() {
    const caja = el('div', '', 'bitacora');
    const escribir = (mensaje, clase) => {
        caja.appendChild(el('div', `[${new Date().toLocaleTimeString()}] ${mensaje}`, clase));
        caja.scrollTop = caja.scrollHeight;
        console.log('[pago]', mensaje);        // eco del cliente
    };
    return { caja, escribir };
}

function campo(etiqueta, valor) {
    const fila = el('div', '', 'campo');
    const input = document.createElement('input');
    input.type = 'text';
    input.value = valor;
    const label = el('label', etiqueta);
    label.appendChild(input);
    fila.appendChild(label);
    return { fila, input };
}

function pintarResultado(caja, r) {
    caja.className = 'resultado-pago';
    if (r === null) {
        caja.textContent = '❌ No fue posible acceder al servidor para realizar el pago';
        caja.classList.add('error-pantalla');
        return;
    }
    const d = r.datos;
    if (r.estado >= 400) {
        caja.textContent = `❌ ${d.error || 'Error ' + r.estado}`;
        caja.classList.add('error-pantalla');
    } else if (d.transaccion_realizada) {
        const cuando = r.intento === 0 ? 'primer intento' : `reintento ${r.intento}`;
        caja.textContent = `✅ ${d.mensaje}. Transacción ${d.id_transaccion} por $${d.monto_transaccion} (obtenido en el ${cuando}).`;
    } else if (d.previamente_realizada) {
        caja.textContent = `⚠️ Este token ya había sido utilizado. Transacción previa ${d.id_transaccion} por $${d.monto_transaccion_previa}. No se cobró de nuevo.`;
    }
}

async function mostrarPlanes() {
    try {
        const planes = await apiGet('/api/suscripciones/planes');
        const lista = el('ul', '', 'menu');
        planes.forEach(p => {
            const li = el('li');
            li.appendChild(el('strong', `${p.nombre}: $${p.precio.toLocaleString('es-CO')} ${p.moneda}`));
            li.appendChild(el('p', `${p.descripcion} (${p.cuentas} cuenta${p.cuentas === 1 ? '' : 's'})`));
            li.appendChild(boton('Elegir', () => mostrarFormularioPago(p)));
            lista.appendChild(li);
        });
        mostrar(el('h2', 'Planes de suscripción'), lista, boton('← Atrás', mostrarTipos));
    } catch (e) { mostrarError(e); }
}

async function mostrarFormularioPago(plan) {
    let simulacion = '';
    try {
        const e = await apiGet('/api/suscripciones/estado');
        simulacion = `Simulación de fallos del servidor: ${e.modo_fallo}` +
            (e.modo_fallo !== 'NINGUNO' ? ` (${e.cantidad_fallos} por token)` : '');
    } catch { /* solo informativo */ }

    const cedula = campo('Cédula', '1061000000');
    const titular = campo('Titular', 'Estudiante Sistemas Distribuidos');
    const tarjeta = campo('Tarjeta', '4111111111111111');
    const bit = crearBitacora();
    const resultado = el('div');

    const btnPagar = boton('💳 Pagar', () => pagar());
    const btnReusar = boton('♻ Reenviar pago con el último token', () => reutilizar());
    btnReusar.disabled = !leerUltimoToken();

    // Mientras hay un pago en curso se bloquean los botones: pulsar "Pagar" dos veces
    // generaría DOS tokens y por tanto DOS cobros. La idempotencia protege los
    // reintentos de un mismo token, no a un usuario que inicia dos pagos.
    const ocupado = (si) => { btnPagar.disabled = si; btnReusar.disabled = si || !leerUltimoToken(); };

    async function pagar() {
        ocupado(true);
        bit.caja.replaceChildren();
        resultado.replaceChildren();
        try {
            bit.escribir('Solicitando token de idempotencia...', '');
            const t = await postJSON('/api/suscripciones/tokens', {
                cedula: cedula.input.value.trim(),
                suscripcion_id: plan.id,
                monto: plan.precio,
                nombre_titular: titular.input.value.trim(),
                numero_tarjeta: tarjeta.input.value.trim()
            }, 5000);

            if (t.estado !== 201) {                      // validación fallida: no hay token
                pintarResultado(resultado, t);
                return;
            }
            const token = t.datos.token_idempotencia;
            guardarUltimoToken({ token, monto: plan.precio, plan: plan.nombre });
            bit.escribir(`Token recibido: ${token}`, 'success');

            pintarResultado(resultado, await enviarPagoConReintentos(token, plan.precio, bit.escribir));
        } catch (e) {
            resultado.textContent = '❌ No fue posible acceder al servidor para obtener el token';
        } finally {
            ocupado(false);
        }
    }

    async function reutilizar() {
        const ultimo = leerUltimoToken();
        if (!ultimo) return;
        ocupado(true);
        bit.caja.replaceChildren();
        resultado.replaceChildren();
        bit.escribir(`Reutilizando el token anterior (plan ${ultimo.plan})`, '');
        pintarResultado(resultado, await enviarPagoConReintentos(ultimo.token, ultimo.monto, bit.escribir));
        ocupado(false);
    }

    mostrar(
        el('h2', `Pagar: ${plan.nombre} ($${plan.precio.toLocaleString('es-CO')})`),
        el('p', simulacion, 'nota'),
        cedula.fila, titular.fila, tarjeta.fila,
        btnPagar, btnReusar,
        resultado, bit.caja,
        boton('← Planes', mostrarPlanes)
    );
}