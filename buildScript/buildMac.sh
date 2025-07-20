#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

# Source the .env file at the beginning of your script
if [ -f ".env" ]; then
    source .env
else
    echo "Error: .env file not found"
    exit 1
fi

# Create build directory if it doesn't exist
if [ ! -d "build" ]; then
  mkdir -p build
fi

# Create dist directory for output files
if [ ! -d "dist" ]; then
  mkdir -p dist
fi

# Copy appicon.png to build directory
cp appicon.png build/

wails build -clean

# --- Configuration Variables ---
WAILS_APP_PATH="./build/bin/Code Zone.app"
SIGNING_IDENTITY="Developer ID Application: ${SIGNING_IDENTITY_LONG}"
APP_NAME=$(basename "$WAILS_APP_PATH" .app)

# Get App Version from the app's Info.plist for DMG naming
APP_VERSION=$(defaults read "$WAILS_APP_PATH/Contents/Info.plist" CFBundleShortVersionString || echo "1.0.0")
DMG_FINAL_PATH="./codezone-${APP_VERSION}_arm64.dmg"

echo "--- Starting Signing and DMG Creation ---"

# --- Step 1: Validate Prerequisites ---
echo "Validating prerequisites..."
[ ! -d "$WAILS_APP_PATH" ] && echo "Error: Wails App bundle not found at '$WAILS_APP_PATH'." && exit 1
[ -z "$SIGNING_IDENTITY" ] || [[ "$SIGNING_IDENTITY" == *"YOUR_TEAM_ID"* ]] && echo "Error: SIGNING_IDENTITY not set correctly." && exit 1

# Check for required command-line tools
command -v codesign >/dev/null 2>&1 || { echo >&2 "Error: codesign not found."; exit 1; }
command -v create-dmg >/dev/null 2>&1 || { echo >&2 "Error: create-dmg not found. Install with: brew install create-dmg"; exit 1; }

echo "Prerequisites validated."

# Sign the application bundle with hardened runtime
echo "Signing the application bundle with hardened runtime..."
codesign --force --deep --sign "${SIGNING_IDENTITY}" --timestamp --options runtime "${WAILS_APP_PATH}"
echo "App bundle signed with hardened runtime."

# Create a pkg for notarization
# echo "Creating pkg for notarization..."
# PKG_PATH="./build/${APP_NAME}.pkg"
# pkgbuild --component "${WAILS_APP_PATH}" --install-location "/Applications" "${PKG_PATH}"

# # Sign the pkg with installer certificate
# echo "Signing the pkg..."
# INSTALLER_IDENTITY="Developer ID Installer: ${SIGNING_IDENTITY_LONG}"
# productsign --sign "${INSTALLER_IDENTITY}" "${PKG_PATH}" "${PKG_PATH}.signed"
# mv "${PKG_PATH}.signed" "${PKG_PATH}"

# # Notarize the pkg
# echo "Notarizing the pkg..."
# xcrun notarytool submit "${PKG_PATH}" \
#   --wait \
#   --apple-id "${NOTARY_USERNAME}" \
#   --password "${NOTARY_PASS}" \
#   --team-id "${SIGNING_IDENTITY_SHORT}"
# echo "Pkg notarized."

# # Staple the notarization ticket to the app
# echo "Stapling notarization ticket..."
# xcrun stapler staple "${WAILS_APP_PATH}"
# echo "Notarization ticket stapled."

# --- Step 3: Create Professional DMG with create-dmg ---
echo "Creating professional DMG file..."

# Remove existing DMG if it exists
rm -f "$DMG_FINAL_PATH"

# Create DMG using create-dmg with smaller icons
echo "Creating DMG with create-dmg..."
create-dmg \
    --volname "$APP_NAME" \
    --window-pos 200 120 \
    --window-size 500 300 \
    --icon-size 64 \
    --icon "$APP_NAME.app" 150 100 \
    --hide-extension "$APP_NAME.app" \
    --app-drop-link 350 100 \
    "$DMG_FINAL_PATH" \
    "$WAILS_APP_PATH"

echo "--- Process Completed ---"
echo "Signed application (.app) is located at: $WAILS_APP_PATH"
echo "DMG file created at: $DMG_FINAL_PATH"

exit 0