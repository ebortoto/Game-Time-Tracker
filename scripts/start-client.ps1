param(
  [switch]$TUI
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

if (Test-Path ".env") {
  Get-Content ".env" | ForEach-Object {
    if ($_ -match '^\s*#' -or $_ -match '^\s*$') { return }
    $parts = $_.Split('=', 2)
    if ($parts.Count -eq 2) {
      [System.Environment]::SetEnvironmentVariable($parts[0].Trim(), $parts[1].Trim())
    }
  }
}

$serverUrl = if ($env:TRACKER_SERVER_URL) { $env:TRACKER_SERVER_URL } else { "http://localhost:8080" }
$apiKey = if ($env:TRACKER_API_KEY) { $env:TRACKER_API_KEY } else { "change-me" }
$configPath = if ($env:TRACKER_CLIENT_CONFIG) { $env:TRACKER_CLIENT_CONFIG } else { "config.json" }
$overlay = if ($env:TRACKER_OVERLAY) { $env:TRACKER_OVERLAY } else { "true" }
$startHidden = if ($env:TRACKER_START_HIDDEN) { $env:TRACKER_START_HIDDEN } else { "true" }

$clientBin = if ($TUI) { ".\bin\tracker-client-cli.exe" } else { ".\bin\tracker-client.exe" }
if (-not (Test-Path $clientBin)) {
  if ($TUI) {
    Write-Host "TUI binary not found. Building console client..."
    go build -o .\bin\tracker-client-cli.exe .\cmd\client
  } else {
    Write-Host "Client binary not found. Building GUI client..."
    go build -ldflags "-H=windowsgui" -o .\bin\tracker-client.exe .\cmd\client
  }
}

$args = @(
  "-config=$configPath"
  "-overlay=$overlay"
  "-start-hidden=$startHidden"
  "-server-url=$serverUrl"
)
if ($apiKey) {
  $args += "-api-key=$apiKey"
}

if ($TUI) {
  if ($PSVersionTable.PSVersion.Major -ge 7) {
    $PSStyle.OutputRendering = "Ansi"
  }
  $args = $args | Where-Object { $_ -notlike "-start-hidden=*" }
  $args += "-start-hidden=false"
  Write-Host "Starting client in interactive TUI mode (PowerShell colors preserved)..."
  & $clientBin @args
  exit $LASTEXITCODE
}

$existing = Get-Process tracker-client -ErrorAction SilentlyContinue
if ($existing) {
  Write-Host "Client is already running (PID $($existing.Id))."
  exit 0
}

$proc = Start-Process -FilePath $clientBin -ArgumentList $args -PassThru
Start-Sleep -Seconds 2

if ($proc.HasExited) {
  Write-Host "Client exited immediately (code $($proc.ExitCode))."
  Write-Host "Try foreground diagnostics to see exact error:"
  Write-Host "  .\scripts\start-client.ps1 -TUI"
  Write-Host "Common cause: TRACKER_API_KEY mismatch between client and server."
  exit 1
}

Write-Host "Client started in background (PID $($proc.Id))."
