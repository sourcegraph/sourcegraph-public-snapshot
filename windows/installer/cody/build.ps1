$CFG = "Release"
$ARCH = "x64"

$ROOT_DIR = "..\..\.."
$BIN_DIR = "${ROOT_DIR}\src-tauri\target\release"
$SIGNED_DIR = "bin\${ARCH}\signed"
$INSTALLER_OUTPUT = "bin\${ARCH}\${CFG}\en-US"
$DIST_DIR_NAME="win-msi"
$DIST_DIR = "${ROOT_DIR}\${DIST_DIR_NAME}"

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

function Write-Banner {
    param(
        [string] $Msg
    )

    if ($env:CI -eq "true") {
        Write-Host "--- ${Msg}"
    } else {
        Write-Host
        Write-Host "--------------------------------------------"
        Write-Host "${Msg}"
    }
}

Write-Host "Building version: ${VERSION}"

Write-Banner -Msg "Cleaning up artifacts"

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

Write-Banner -Msg "Preparing artifacts for build"

New-Item -ItemType Directory "${SIGNED_DIR}"

Copy-Item -Path "${BIN_DIR}\Cody.exe" -Destination "${SIGNED_DIR}\Cody.exe"
Copy-Item -Path "${BIN_DIR}\sourcegraph-backend.exe" -Destination "${SIGNED_DIR}\sourcegraph-backend.exe"
Start-Sleep -Seconds 1 # BUG: Powershell is leaving files open after copy and fails signing.

Write-Banner -Msg "Signing artifacts"

./sign.ps1 "${SIGNED_DIR}\Cody.exe"
./sign.ps1 "${SIGNED_DIR}\sourcegraph-backend.exe"

Write-Banner -Msg "Building installer for ${ARCH} ${CFG}"

msbuild /restore /p:Configuration=${CFG} /p:Platform=${ARCH}

Write-Banner -Msg "Signing installer"

$MSI_PATH = "${INSTALLER_OUTPUT}\cody-${VERSION}-${ARCH}.msi"
./sign.ps1 $MSI_PATH

# Only upload if we're in CI
if ($env:CI -eq "true" ) {
    Write-Banner -Msg "Uploading ${MSI_PATH}"

    New-Item -Path "${DIST_DIR}" -ItemType Directory -Force
    Write-Host "Moving ${MSI_PATH} to ${DIST_DIR}"
    Copy-Item -Path "${MSI_PATH}" -Destination ${DIST_DIR}
    Start-Sleep -Seconds 1 # BUG: Powershell is leaving files open after copy and fails signing.
    Resolve-Path -Path ${ROOT_DIR}
    Push-Location ${ROOT_DIR}
    Write-Host "Uploading artifacts from ${DIST_DIR}"
    buildkite-agent artifact upload "${DIST_DIR_NAME}/*"
    Pop-Location
}
