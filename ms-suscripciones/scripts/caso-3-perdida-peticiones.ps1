# Caso 3: sin respuesta tras el cuarto reintento por perdida de mensajes de peticion
# MODO_FALLO=PETICION  CANTIDAD_FALLOS=5
$env:MODO_FALLO = "PETICION"
$env:CANTIDAD_FALLOS = "5"
$env:RETARDO_FALLO_MS = "5000"
Set-Location (Join-Path $PSScriptRoot "..")
go run .
