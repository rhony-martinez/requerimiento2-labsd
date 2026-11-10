# Caso 2: transaccion realizada al tercer reintento (se pierden 3 peticiones)
# MODO_FALLO=PETICION  CANTIDAD_FALLOS=3
$env:MODO_FALLO = "PETICION"
$env:CANTIDAD_FALLOS = "3"
$env:RETARDO_FALLO_MS = "5000"
Set-Location (Join-Path $PSScriptRoot "..")
go run .
