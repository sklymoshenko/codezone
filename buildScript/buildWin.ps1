# PowerShell build script for Code Zone
# Exit immediately if a command exits with a non-zero status.
$ErrorActionPreference = "Stop"

# Create build directory if it doesn't exist
if (-not (Test-Path "build")) {
    New-Item -ItemType Directory -Path "build" -Force
}


# Copy appicon.png to build directory
Copy-Item "appicon.png" "build/" -Force

# Run wails build with UPX compression
wails build -clean -nsis -upx -upxflags --lzma

# Get app version from package.json in the frontend folder
$appVersion = ""
$packageJsonPath = "frontend\package.json"
if (Test-Path $packageJsonPath) {
    $packageJson = Get-Content $packageJsonPath -Raw | ConvertFrom-Json
    $appVersion = $packageJson.version
} else {
    Write-Error "frontend/package.json not found. Cannot determine app version."
    exit 1
}

# Get architecture
$arch = if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") { "x64" } else { $env:PROCESSOR_ARCHITECTURE.ToLower() }

# Find the built .exe files in the build directory
$exes = Get-ChildItem -Path "build/bin" -Filter "*.exe"
$baseName = "codezone"

if ($exes.Count -gt 1) {
    # Compare file sizes and add nsis_ prefix to the bigger one
    $sortedExes = $exes | Sort-Object Length -Descending
    $biggestExe = $sortedExes[0]
    $otherExes = $sortedExes | Select-Object -Skip 1

    # Rename the biggest exe with nsis_ prefix
    $exeNewName = "nsis_${baseName}_${appVersion}_${arch}.exe"
    Rename-Item -Path $biggestExe.FullName -NewName $exeNewName -Force

    # Rename the rest normally
    foreach ($exe in $otherExes) {
        $exeNewName = "${baseName}_${appVersion}_${arch}.exe"
        Rename-Item -Path $exe.FullName -NewName $exeNewName -Force
    }
} else {
    foreach ($exe in $exes) {
        $exeNewName = "${baseName}_${appVersion}_${arch}.exe"
        Rename-Item -Path $exe.FullName -NewName $exeNewName -Force
    }
}

Write-Host "Build finished successfully." -ForegroundColor Green
