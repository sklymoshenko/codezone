#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

# --- Configuration ---
APP_NAME="codezone"
VERSION="1.0.0"
ARCH="amd64"
MAINTAINER="Stanislav Klymoshenko"
DESCRIPTION="Code Zone - A desktop code playground for JavaScript, Go, and SQL."

# --- Script ---

echo "--- Starting Linux Debian Package Build ---"

# 1. Build the Wails application
echo "--> Building Wails binary..."
wails build -upx -upxflags --lzma

# 2. Set up packaging directory
echo "--> Creating packaging directory structure..."
rm -rf packaging # Clean up old structure
mkdir -p packaging/DEBIAN
mkdir -p packaging/opt/$APP_NAME
mkdir -p packaging/usr/share/applications
mkdir -p packaging/usr/share/icons/hicolor/128x128/apps

# 3. Copy files into the packaging directory
echo "--> Copying application files..."
cp ./build/bin/$APP_NAME packaging/opt/$APP_NAME/
cp ./appicon.png packaging/usr/share/icons/hicolor/128x128/apps/$APP_NAME.png
cp ./codezone.desktop packaging/usr/share/applications/

# 4. Create the Debian control file
echo "--> Creating DEBIAN/control file..."
cat > packaging/DEBIAN/control << EOF
Package: $APP_NAME
Version: $VERSION
Architecture: $ARCH
Maintainer: $MAINTAINER
Description: $DESCRIPTION
Depends: libwebkit2gtk-4.0-37, libgtk-3-0
EOF

# 5. Build the .deb package
echo "--> Building .deb package..."
DEB_NAME="${APP_NAME}_${VERSION}_${ARCH}.deb"
dpkg-deb --build --root-owner-group packaging ${DEB_NAME}

# 6. Clean up temporary directories
echo "--> Cleaning up..."
rm -rf build
rm -rf packaging

echo "--- Build Complete ---"
echo "Package created: ${DEB_NAME}"

# Send the filename to GitHub Actions
if [ -n "$GITHUB_OUTPUT" ]; then
  echo "deb_path=${DEB_NAME}" >> "$GITHUB_OUTPUT"
fi 