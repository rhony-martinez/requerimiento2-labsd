// Subida con progreso. Resuelve {estado, cuerpo}; rechaza solo si no hubo respuesta (red caída).
function subirConProgreso(url, form, cabeceras, alProgreso) {
  return new Promise((resolver, rechazar) => {
    const xhr = new XMLHttpRequest();
    xhr.open("POST", url);
    Object.entries(cabeceras).forEach(([k, v]) => xhr.setRequestHeader(k, v));
    // SIN Content-Type: el navegador pone el boundary del multipart.

    xhr.upload.onprogress = (e) => {
      if (e.lengthComputable) alProgreso(e.loaded / e.total);
    };
    xhr.onload = () => {
      let cuerpo = {};
      try {
        cuerpo = JSON.parse(xhr.responseText);
      } catch {
        /* respuesta sin JSON */
      }
      resolver({ estado: xhr.status, cuerpo });
    };
    xhr.onerror = () => rechazar(new Error("red"));
    xhr.send(form);
  });
}

// Vista de administración: registrar un audio nuevo (multipart/form-data).

const MAX_MB = 50;

// Lee la duración del mp3 con un <audio> en memoria, para autocompletar el campo.
// function leerDuracion(archivo) {
//     return new Promise(resolver => {
//         const url = URL.createObjectURL(archivo);
//         const a = new Audio();
//         a.preload = 'metadata';
//         a.onloadedmetadata = () => {
//             URL.revokeObjectURL(url);
//             resolver(Number.isFinite(a.duration) ? Math.round(a.duration) : null);
//         };
//         a.onerror = () => { URL.revokeObjectURL(url); resolver(null); };
//         a.src = url;
//     });
// }

async function mostrarAdmin() {
  try {
    const tipos = await apiGet("/api/catalogo/tipos");

    const clave = campo(
      "Clave de administrador",
      sessionStorage.getItem("claveAdmin") || "",
    );
    clave.input.type = "password";

    const selector = document.createElement("select");
    tipos.forEach((t) => {
      const o = document.createElement("option");
      o.value = t.id;
      o.textContent = t.nombre;
      selector.appendChild(o);
    });
    const filaTipo = el("div", "", "campo");
    const etiquetaTipo = el("label", "Tipo de audio");
    etiquetaTipo.appendChild(selector);
    filaTipo.appendChild(etiquetaTipo);

    const titulo = campo("Título", "");
    // const duracion = campo('Duración (segundos)', '');
    // duracion.input.type = 'number';
    // duracion.input.min = '1';

    const entradaArchivo = document.createElement("input");
    entradaArchivo.type = "file";
    entradaArchivo.accept = ".mp3,audio/mpeg";
    const filaArchivo = el("div", "", "campo");
    const etiquetaArchivo = el("label", "Archivo mp3");
    etiquetaArchivo.appendChild(entradaArchivo);
    filaArchivo.appendChild(etiquetaArchivo);

    const zonaCampos = el("div"); // metadatos propios del tipo
    let camposActuales = []; // [{clave, input}]
    const mensaje = el("p", "", "resultado-pago");

    function avisar(texto, esError) {
      mensaje.textContent = texto;
      mensaje.className = "resultado-pago" + (esError ? " error-pantalla" : "");
    }

    // Los campos los dicta el servidor: este cliente no conoce los metadatos de cada tipo.
    async function cargarCampos() {
      try {
        const defs = await apiGet(
          `/api/catalogo/tipos/${selector.value}/campos`,
        );
        camposActuales = defs.map((d) => ({
          clave: d.clave,
          ...campo(d.etiqueta, ""),
        }));
        zonaCampos.replaceChildren(...camposActuales.map((c) => c.fila));
      } catch (e) {
        avisar(e.message, true);
      }
    }
    selector.addEventListener("change", cargarCampos);

    // entradaArchivo.addEventListener('change', async () => {
    //     const f = entradaArchivo.files[0];
    //     if (!f) return;
    //     const seg = await leerDuracion(f);
    //     if (seg) duracion.input.value = seg;   // editable por si acaso
    // });

    const barra = document.createElement("progress");
    barra.max = 100;
    barra.value = 0;
    barra.hidden = true;
    const btnRegistrar = boton("⬆ Registrar audio", registrar);

    async function registrar() {
      const archivo = entradaArchivo.files[0];
      if (!archivo) {
        avisar("Selecciona un archivo mp3.", true);
        return;
      }
      if (archivo.size > MAX_MB * 1024 * 1024) {
        avisar(`El archivo supera ${MAX_MB} MB.`, true);
        return;
      }
      const metadatos = {};
      camposActuales.forEach((c) => {
        metadatos[c.clave] = c.input.value.trim();
      });

      const form = new FormData();
      form.append("tipo_id", selector.value);
      form.append("titulo", titulo.input.value.trim());
      // form.append('duracion_seg', duracion.input.value);
      form.append("metadatos", JSON.stringify(metadatos));
      form.append("archivo", archivo); // el archivo al final

      barra.value = 0;
      barra.hidden = false;
      btnRegistrar.disabled = true;
      avisar("Subiendo...", false);
      try {
        sessionStorage.setItem("claveAdmin", clave.input.value);
      } catch {
        /* opcional */
      }

      try {
        const r = await subirConProgreso(
          `${window.GATEWAY_URL}/api/catalogo/audios`,
          form,
          { "X-Admin-Token": clave.input.value },
          (fraccion) => {
            barra.value = Math.round(fraccion * 100);
            avisar(
              fraccion < 1
                ? `Subiendo... ${barra.value}%`
                : "Archivo enviado; el servidor lo está procesando...",
              false,
            );
          },
        );

        if (r.estado >= 400) {
          avisar(`❌ ${r.cuerpo.error || "Error HTTP " + r.estado}`, true);
        } else {
          avisar(
            `✅ Audio registrado con id ${r.cuerpo.id}: ${r.cuerpo.titulo} (${formatearDuracion(r.cuerpo.duracion_seg)})`,
            false,
          );
          titulo.input.value = "";
          entradaArchivo.value = "";
          camposActuales.forEach((c) => {
            c.input.value = "";
          });
        }
      } catch {
        avisar(
          "❌ No se pudo contactar al API Gateway. ¿Está encendido?",
          true,
        );
      } finally {
        barra.hidden = true;
        btnRegistrar.disabled = false;
      }
    }

    mostrar(
      el("h2", "Administración: registrar audio"),
      clave.fila,
      filaTipo,
      titulo.fila,
      zonaCampos,
      filaArchivo,
      btnRegistrar,
      barra,
      mensaje,
      boton("🗑 Gestionar audios", mostrarGestionAudios),
      boton("← Atrás", mostrarTipos),
    );
    cargarCampos();
  } catch (e) {
    mostrarError(e);
  }
}

async function mostrarGestionAudios() {
  try {
    const tipos = await apiGet("/api/catalogo/tipos");
    const clave = campo(
      "Clave de administrador",
      sessionStorage.getItem("claveAdmin") || "",
    );
    clave.input.type = "password";
    const salida = el("div");
    const mensaje = el("p", "", "resultado-pago");

    async function recargar() {
      const bloques = [];
      for (const t of tipos) {
        const audios = await apiGet(`/api/catalogo/tipos/${t.id}/audios`);
        bloques.push(el("h3", `${t.nombre} (${audios.length})`));
        if (audios.length === 0) bloques.push(el("p", "Sin audios.", "nota"));
        audios.forEach((a) => {
          const fila = el("div", "", "campo");
          fila.appendChild(el("span", `#${a.id}  ${a.titulo}  `));
          fila.appendChild(boton("🗑 Eliminar", () => eliminar(a)));
          bloques.push(fila);
        });
      }
      salida.replaceChildren(...bloques);
    }

    async function eliminar(a) {
      if (
        !confirm(`¿Eliminar "${a.titulo}" (id ${a.id})? No se puede deshacer.`)
      )
        return;
      try {
        const resp = await fetch(
          `${window.GATEWAY_URL}/api/catalogo/audios/${a.id}`,
          {
            method: "DELETE",
            headers: { "X-Admin-Token": clave.input.value },
          },
        );
        if (!resp.ok) {
          const c = await resp.json().catch(() => ({}));
          throw new Error(c.error || `Error HTTP ${resp.status}`);
        }
        mensaje.className = "resultado-pago";
        mensaje.textContent = `✅ Audio ${a.id} eliminado`;
        await recargar();
      } catch (e) {
        mensaje.className = "resultado-pago error-pantalla";
        mensaje.textContent = `❌ ${e.message}`;
      }
    }

    mostrar(
      el("h2", "Administración: gestionar audios"),
      clave.fila,
      mensaje,
      salida,
      boton("← Atrás", mostrarAdmin),
    );
    await recargar();
  } catch (e) {
    mostrarError(e);
  }
}
