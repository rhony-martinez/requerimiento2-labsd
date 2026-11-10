# Borra el historial de tokens y transacciones para empezar la practica desde cero.
$datos = Join-Path $PSScriptRoot "..\datos"
if (Test-Path $datos) { Remove-Item -Recurse -Force $datos }
Write-Host "Historial de tokens y transacciones eliminado."
