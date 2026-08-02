$ErrorActionPreference = "Stop"

$installDir = Join-Path $env:LOCALAPPDATA "Programs\chapteriser"
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
$segments = @($userPath -split ';' | Where-Object { $_ -and $_ -ne $installDir })
[Environment]::SetEnvironmentVariable("Path", ($segments -join ';'), "User")
Remove-Item -Recurse -Force -ErrorAction SilentlyContinue $installDir
Write-Output "chapteriser has been removed."
