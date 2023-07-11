# Get the current script path
$currentScriptPath = $PSScriptRoot

# Move three directories back
$targetPath = (Split-Path -Path $currentScriptPath -Parent)
$targetPath = (Split-Path -Path $targetPath -Parent)
$targetPath = (Split-Path -Path $targetPath -Parent)

$ROOT_DIR = $targetPath

class Version {
    [int] $major
    [int] $minor
    [int] $build
    [string] $rev

    [string] String(){
        return "$($this.major).$($this.minor).$($this.build).$($this.rev)"
    }
}

function Create-Version {
    $version = [Version]::new()
# Core version
    $version.major = 5
    $version.minor = 1
    $version.build = 0

# Creates the BUILD version
    $epoch =  Get-Date -Year 2023 -Month 05 -Day 22 -Hour 08 -Minute 00 -Second 00
    $now = Get-Date
    $version.rev = [math]::Round(($now - $epoch).TotalHours * 4)

    return $version
}

function Update-Tauri-Conf-Version {
    param(
        [string] $Version
    )

    $configPath = "${ROOT_DIR}/src-tauri/tauri.conf.json"

    if (Test-Path -Path "${configPath}" -eq $false) {
        throw "Cannot update version. Failed to find tauri config file at ${configPath}"
    }
    $conf = Get-Content -Raw "$configPath" | ConvertFrom-Json
    $conf.package.version = "$Version"

    $conf | ConvertTo-Json | Out-File "$configPath"
}

$version = Create-Version

if ($env:CI -eq "true") {
   Update-Tauri-Conf-Version $version.String()
}

Write-Host $version.String()
