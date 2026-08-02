param(
    [Parameter(Mandatory = $true)][string]$Version,
    [Parameter(Mandatory = $true)][string]$VoskDir,
    [Parameter(Mandatory = $true)][string]$FFmpegDir,
    [Parameter(Mandatory = $true)][string]$OutputDir
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$bundle = "chapteriser_${Version}_windows_amd64"
$stage = Join-Path $OutputDir $bundle

foreach ($required in @(
    (Join-Path $VoskDir "libvosk.dll"),
    (Join-Path $FFmpegDir "ffmpeg.exe"),
    (Join-Path $FFmpegDir "ffprobe.exe")
)) {
    if (-not (Test-Path -LiteralPath $required -PathType Leaf)) {
        throw "Missing required release dependency: $required"
    }
}

New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null
Remove-Item -Recurse -Force -ErrorAction SilentlyContinue $stage
New-Item -ItemType Directory -Force -Path $stage | Out-Null

Push-Location $root
try {
    $ldflags = "-s -w -X main.version=$Version"
    go build -trimpath -ldflags $ldflags -o (Join-Path $stage "chapteriser.exe") ./cmd/chapteriser
} finally {
    Pop-Location
}

Copy-Item (Join-Path $FFmpegDir "ffmpeg.exe"), (Join-Path $FFmpegDir "ffprobe.exe") -Destination $stage
Get-ChildItem -Path $VoskDir -Filter "*.dll" -File | Copy-Item -Destination $stage
Copy-Item -Recurse (Join-Path $root "model") (Join-Path $stage "model")
Copy-Item (Join-Path $root "README.md"), (Join-Path $root "packaging\\windows\\install.ps1"), (Join-Path $root "packaging\\windows\\uninstall.ps1") -Destination $stage

$archive = Join-Path $OutputDir "$bundle.zip"
Remove-Item -Force -ErrorAction SilentlyContinue $archive
Compress-Archive -Path $stage -DestinationPath $archive
Write-Output "Created $archive"
