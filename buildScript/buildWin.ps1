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
wails build -upx -upxflags --lzma