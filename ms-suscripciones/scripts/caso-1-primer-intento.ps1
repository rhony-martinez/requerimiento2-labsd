# Caso 1: transaccion realizada al primer intento
# MODO_FALLO=NINGUNO  CANTIDAD_FALLOS=0
$env:MODO_FALLO = "NINGUNO"
$env:CANTIDAD_FALLOS = "0"
$env:RETARDO_FALLO_MS = "5000"
Set-Location (Join-Path $PSScriptRoot "..")
go run .
