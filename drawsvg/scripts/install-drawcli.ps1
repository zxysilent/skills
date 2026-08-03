param(
    [switch]$Force
)

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$baseUrl = if ($env:DRAWCLI_BASE_URL) {
    $env:DRAWCLI_BASE_URL
} else {
    "https://raw.githubusercontent.com/zxysilent/skills/master/bins"
}

$arch = if ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture -eq "Arm64") {
    "arm64"
} else {
    "amd64"
}
$name = "drawcli-windows-$arch.exe"
$url = "$baseUrl/$name"
$output = Join-Path $scriptDir "drawcli.exe"

if (-not $Force -and (Test-Path $output)) {
    Write-Host "drawcli already installed at $output (use -Force to reinstall)"
    return
}

Invoke-WebRequest -Uri $url -OutFile $output
Write-Host "installed $url -> $output"
