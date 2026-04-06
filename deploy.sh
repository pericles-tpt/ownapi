#!/bin/sh

# Read environment variables
source 'build.env'

# Clear out old `build/` directory
setopt rmstarsilent
rm -rf build/
mkdir build/

# Build frontend bundle
if [ ! -d "./node_modules" ]; then
    chmod +x ../unpack.sh
    npm i
fi
npx webpack && cd ..

# Build frontend Go code
GOOS=linux GOARCH=amd64 go build -o ./build/projected

# Copy required files to `build/` directory
cp -R config.json build/
mkdir -p build/_frontend/dist/
cp dist/*.js build/_frontend/dist/
mkdir -p build/_frontend/src/
cp src/*.html build/_frontend/src/
cp .env build/.env
cp unpack.sh build/unpack.sh
mkdir build/generated

# Zip everything in build
# NOTE: Excludes `tables.json` for now, since that's modified on the server (need to figure out how to "sync" those changes)
cd build && zip -r build.zip generated/ config.json dist src/*.html .env projected unpack.sh

# Send to VPS
scp -r -P $VPS_PORT ./build.zip "${SCP_DST_USER}@${SCP_DST_HOST}:${SCP_DST_PATH}/build.zip"