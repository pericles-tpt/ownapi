#!/bin/sh

# Clear out old `build/` directory
setopt rmstarsilent
rm -rf build/
mkdir build/

# Build frontend bundle
if [ ! -d "./node_modules" ]; then
    npm i
fi
npx webpack

# WARNING: Any flags that modify the output binary here MUST
#          also be added to the plugin build step in the code
GOOS=linux GOARCH=amd64 go build -o ./build/

# Copy required files to `build/` directory
mkdir -p build/_bin
cp -R _bin/* build/_bin/
mkdir -p build/_config
cp -R _config/* build/_config/
mkdir -p build/_functions
cp -R _functions/* build/_functions/
mkdir -p build/_data
cp -R _data/* build/_data
mkdir -p build/dist
cp -R dist/* build/dist
mkdir -p build/node_modules
cp -R node_modules/* build/node_modules
mkdir -p build/src
cp -R src/* build/src
mkdir -p build/user_functions
cp -R user_functions/* build/user_functions
cp *.json build/
cp README.md build/
cp secrets.txt build/
cp tsconfig.json build/
cp webpack.config.js build/
cp .env build/
cp ~/Desktop/thing.txt build/

# Zip everything in build
cd build && zip -r ../ownapi.zip *
cd ..