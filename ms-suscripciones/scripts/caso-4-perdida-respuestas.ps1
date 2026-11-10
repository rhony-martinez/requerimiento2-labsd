# Caso 4: sin respuesta tras el cuarto reintento por perdida de mensajes de respuesta
# MODO_FALLO=RESPUESTA  CANTIDAD_FALLOS=5
$env:MODO_FALLO = "RESPUESTA"
$env:CANTIDAD_FALLOS = "5"
$env:RETARDO_FALLO_MS = "5000"
Set-Location (Join-Path $PSScriptRoot "..")
go run .
