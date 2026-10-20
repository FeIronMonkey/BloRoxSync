# One-command rbx-sync install for Windows (windows/amd64 asset).
param(
    [switch]$DryRun
)

$ErrorActionPreference = "Stop"

$AssetBin = "rbx-sync-windows-amd64.exe"
$AssetPlugin = "rbx-sync.rbxmx"
$AssetSums = "SHA256SUMS"
$DestDir = Join-Path $env:LOCALAPPDATA "rbx-sync"
$DestBin = Join-Path $DestDir "rbx-sync.exe"
$DestPlugin = Join-Path $DestDir "rbx-sync.rbxmx"

function Write-DestTable {
    Write-Output "asset: $AssetBin"
    Write-Output "asset: $AssetPlugin"
    Write-Output "asset: $AssetSums"
    Write-Output "dest: $DestBin"
    Write-Output "dest: $DestPlugin"
    Write-Output "Add $DestDir to the user PATH when missing. A new terminal is required."
}

if ($DryRun) {
    Write-DestTable
    exit 0
}

if ([string]::IsNullOrWhiteSpace($env:RBX_SYNC_RELEASE_BASE)) {
    Write-Error "RBX_SYNC_RELEASE_BASE is unset; set it to the HTTPS directory of release assets (see README.md)."
    exit 1
}

$Base = $env:RBX_SYNC_RELEASE_BASE.TrimEnd("/")
$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("rbx-sync-install-" + [guid]::NewGuid().ToString("n"))
New-Item -ItemType Directory -Path $tmp | Out-Null
try {
    function Fetch([string]$name) {
        Invoke-WebRequest -UseBasicParsing -Uri "$Base/$name" -OutFile (Join-Path $tmp $name) | Out-Null
    }
    Fetch $AssetSums
    Fetch $AssetBin
    Fetch $AssetPlugin

    $expected = @{}
    Get-Content -Path (Join-Path $tmp $AssetSums) | ForEach-Object {
        if ($_ -match '^(?<h>[0-9a-fA-F]{64})\s+\*?(?<n>\S+)\s*$') {
            $expected[$Matches.n] = $Matches.h.ToLowerInvariant()
        }
    }

    function Assert-Hash([string]$name) {
        $path = Join-Path $tmp $name
        $got = (Get-FileHash -Algorithm SHA256 -Path $path).Hash.ToLowerInvariant()
        if (-not $expected.ContainsKey($name) -or $expected[$name] -ne $got) {
            throw "SHA-256 mismatch for $name; refusing to replace installed files."
        }
    }
    Assert-Hash $AssetBin
    Assert-Hash $AssetPlugin

    New-Item -ItemType Directory -Path $DestDir -Force | Out-Null
    Copy-Item -Force (Join-Path $tmp $AssetBin) $DestBin
    Copy-Item -Force (Join-Path $tmp $AssetPlugin) $DestPlugin
}
finally {
    Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}

$canonicalDest = Join-Path ([Environment]::GetFolderPath("LocalApplicationData")) "rbx-sync"
if ($DestDir -eq $canonicalDest) {
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($null -eq $userPath) { $userPath = "" }
    $parts = $userPath.Split(";", [StringSplitOptions]::RemoveEmptyEntries)
    if (-not ($parts -contains $DestDir)) {
        $joined = if ($userPath.Trim() -eq "") { $DestDir } else { "$userPath;$DestDir" }
        [Environment]::SetEnvironmentVariable("Path", $joined, "User")
    }
}
Write-Output "Add $DestDir to the user PATH when missing. A new terminal is required."

$pluginsDir = Join-Path $env:LOCALAPPDATA "Roblox\Plugins"
if (Test-Path -LiteralPath $pluginsDir -PathType Container) {
    & $DestBin plugin install
    exit $LASTEXITCODE
}

Write-Output "staged plugin: $DestPlugin"
exit 0
