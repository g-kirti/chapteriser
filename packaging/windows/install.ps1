$ErrorActionPreference = "Stop"

$sourceDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$installDir = Join-Path $env:LOCALAPPDATA "Programs\chapteriser"
$parent = Split-Path -Parent $installDir
New-Item -ItemType Directory -Force -Path $parent | Out-Null
Remove-Item -Recurse -Force -ErrorAction SilentlyContinue $installDir
New-Item -ItemType Directory -Force -Path $installDir | Out-Null
Copy-Item -Recurse -Force (Join-Path $sourceDir "*") $installDir

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
$segments = @($userPath -split ';' | Where-Object { $_ })
if ($segments -notcontains $installDir) {
    [Environment]::SetEnvironmentVariable("Path", (($segments + $installDir) -join ';'), "User")
}

Write-Output "Installed chapteriser to $installDir. Open a new terminal before running chapteriser."
