$ErrorActionPreference = 'Stop'

$Root = $PSScriptRoot
$Source = Join-Path $Root 'fnpack'
$FnpackName = 'fnpack-1.2.3-linux-amd64'
$GoCommand = Get-Command go.exe -ErrorAction SilentlyContinue
if (-not $GoCommand) { $GoCommand = Get-Command go -ErrorAction SilentlyContinue }
if (-not $GoCommand) { throw 'Go was not found. Install Go and add it to PATH.' }
$WslCommand = Get-Command wsl.exe -ErrorAction SilentlyContinue
if (-not $WslCommand) { throw 'WSL was not found. Install WSL Ubuntu before building.' }

function Invoke-GoBuild([string]$Architecture, [string]$Output) {
  $env:GOOS = 'linux'
  $env:GOARCH = $Architecture
  $env:CGO_ENABLED = '0'
  & $GoCommand.Source build -trimpath -ldflags '-s -w' -o $Output $Root
  if ($LASTEXITCODE -ne 0) { throw "Linux $Architecture backend build failed." }
}

function New-Staging([string]$Name, [string]$Platform, [string]$Architecture, [string]$ExcludedSkill) {
  $stage = Join-Path $Root "fnpack-$Name"
  if (Test-Path -LiteralPath $stage) { Remove-Item -LiteralPath $stage -Recurse -Force }
  & robocopy $Source $stage /E /XF $ExcludedSkill '*.fpk' /NFL /NDL /NJH /NJS /NP | Out-Null
  if ($LASTEXITCODE -ge 8) { throw "$Name staging copy failed." }

  Invoke-GoBuild $Architecture (Join-Path $stage 'app/bin/dsh.tavern')
  if ($Platform -eq 'arm') {
    $manifestPath = Join-Path $stage 'manifest'
    $manifest = [IO.File]::ReadAllText($manifestPath)
    $manifest = [Text.RegularExpressions.Regex]::Replace($manifest, '(?m)^platform\s*=\s*x86\s*$', 'platform              = arm')
    [IO.File]::WriteAllText($manifestPath, $manifest, [Text.UTF8Encoding]::new($false))
  }
  return $stage
}

function Invoke-Fnpack([string]$Stage) {
  $drive = $Stage.Substring(0, 1).ToLowerInvariant()
  $wslStage = "/mnt/$drive" + $Stage.Substring(2).Replace('\', '/')
  $command = "cd '$wslStage' && chmod +x $FnpackName && ./$FnpackName build"
  & $WslCommand.Source -d Ubuntu -- bash -lc $command
  if ($LASTEXITCODE -ne 0) { throw "$Stage package build failed." }
}

Write-Host '[INFO] Building Vue 3 frontend assets...'
Push-Location (Join-Path $Root 'frontend')
try {
  & npm ci
  if ($LASTEXITCODE -ne 0) { throw 'Frontend dependency installation failed.' }
  & npm run build
  if ($LASTEXITCODE -ne 0) { throw 'Frontend build failed.' }
}
finally { Pop-Location }

$TavernSource = Join-Path $Root 'tavern-source'
$TavernArchive = Join-Path $Source 'app/dsh-tavern.tar.gz'
if (-not (Test-Path -LiteralPath $TavernSource -PathType Container)) { throw 'tavern-source directory is missing.' }
Write-Host '[INFO] Packaging embedded DSH Tavern source...'
& tar.exe -czf $TavernArchive -C $TavernSource .
if ($LASTEXITCODE -ne 0) { throw 'Embedded DSH Tavern source packaging failed.' }

$x86Stage = New-Staging 'x86' 'x86' 'amd64' 'trim-cli-linux-arm64'
$armStage = New-Staging 'arm' 'arm' 'arm64' 'trim-cli-linux-x64'
Invoke-Fnpack $x86Stage
Invoke-Fnpack $armStage
Copy-Item (Join-Path $x86Stage 'dsh.tavern.fpk') (Join-Path $Root 'dsh.tavern-x86.fpk') -Force
Copy-Item (Join-Path $armStage 'dsh.tavern.fpk') (Join-Path $Root 'dsh.tavern-arm.fpk') -Force
Write-Host '[INFO] Created dsh.tavern-x86.fpk and dsh.tavern-arm.fpk.'
