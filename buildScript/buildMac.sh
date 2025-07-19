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
DMG_FINAL_PATH="./dist/codezone-${APP_VERSION}_arm64.dmg"

echo "--- Starting Signing and DMG Creation ---"

# --- Step 1: Validate Prerequisites ---
echo "Validating prerequisites..."
[ ! -d "$WAILS_APP_PATH" ] && echo "Error: Wails App bundle not found at '$WAILS_APP_PATH'." && exit 1
[ -z "$SIGNING_IDENTITY" ] || [[ "$SIGNING_IDENTITY" == *"YOUR_TEAM_ID"* ]] && echo "Error: SIGNING_IDENTITY not set correctly." && exit 1

# Check for required command-line tools
command -v codesign >/dev/null 2>&1 || { echo >&2 "Error: codesign not found."; exit 1; }
command -v create-dmg >/dev/null 2>&1 || { echo >&2 "Error: create-dmg not found. Install with: brew install create-dmg"; exit 1; }

echo "Prerequisites validated."

# --- Step 2: Clean and Sign the App Bundle in Place ---
echo "Cleaning and signing the app bundle in place..."

# Clean up any extended attributes
xattr -cr "$WAILS_APP_PATH"

# Set correct permissions
chmod -R 755 "$WAILS_APP_PATH"
find "$WAILS_APP_PATH" -type f -exec chmod 644 {} \;

# Sign the application bundle
echo "Signing the application bundle..."
codesign --force --deep --sign "${SIGNING_IDENTITY}" --timestamp "${WAILS_APP_PATH}"
echo "App bundle signed."

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