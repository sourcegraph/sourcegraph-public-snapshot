$CFG = "Release"
$ARCH = "x64"

$BIN_DIR = "..\..\..\src-tauri\target\release"
$SIGNED_DIR = "bin\${ARCH}\signed"
$INSTALLER_OUTPUT = "bin\${ARCH}\${CFG}\en-US"

# Core version
$major = 5
$minor = 1
$build = 0

# Creates the BUILD version
$epoch =  Get-Date -Year 2023 -Month 05 -Day 22 -Hour 08 -Minute 00 -Second 00
$now = Get-Date
$rev = [math]::Round(($now - $epoch).TotalHours * 4)

$VERSION = "${major}.${minor}.${build}.${rev}"

# Write installer version
$ver = [xml]"<Project xmlns='http://schemas.microsoft.com/developer/msbuild/2003'>
    <PropertyGroup>
        <InstallerMajorVersion>${major}</InstallerMajorVersion>
        <InstallerMinorVersion>${minor}</InstallerMinorVersion>
        <InstallerBuildVersion>${build}</InstallerBuildVersion>
        <InstallerRevVersion>${rev}</InstallerRevVersion>
  </PropertyGroup>
</Project>"
$ver.save("cody.version.inc")

Write-Host "Building version: ${VERSION}"

Write-Host
Write-Host "--------------------------------------------"
Write-Host "Cleaning up artifacts"
Write-Host

if (Test-Path -Path "${SIGNED_DIR}") {
    Remove-Item -Recurse -Force "${SIGNED_DIR}"
    if (Test-Path -Path "${SIGNED_DIR}") {
        throw "Failed to delete signed dir"
    }
}

if (Test-Path -Path "${INSTALLER_OUTPUT}") {
    Remove-Item -Recurse -Force "${INSTALLER_OUTPUT}"
    if (Test-Path -Path "${INSTALLER_OUTPUT}") {
        throw "Failed to delete installer output"
    }
}

Write-Host
Write-Host "--------------------------------------------"
Write-Host "Preparing artifacts for build"
Write-Host

New-Item -ItemType Directory "${SIGNED_DIR}"

Copy-Item -Path "${BIN_DIR}\Cody.exe" -Destination "${SIGNED_DIR}\Cody.exe"
Copy-Item -Path "${BIN_DIR}\sourcegraph-backend.exe" -Destination "${SIGNED_DIR}\sourcegraph-backend.exe"
Start-Sleep -Seconds 1 # BUG: Powershell is leaving files open after copy and fails signing.

Write-Host
Write-Host "--------------------------------------------"
Write-Host "Signing artifacts"
Write-Host

./sign.ps1 "${SIGNED_DIR}\Cody.exe"
./sign.ps1 "${SIGNED_DIR}\sourcegraph-backend.exe"

Write-Host
Write-Host "--------------------------------------------"
Write-Host "Building installer for ${ARCH} ${CFG}"
Write-Host

msbuild /p:Configuration=${CFG} /p:Platform=${ARCH}

Write-Host
Write-Host "--------------------------------------------"
Write-Host "Signing installer"
Write-Host

./sign.ps1 "${INSTALLER_OUTPUT}\cody-${VERSION}-${ARCH}.msi"
