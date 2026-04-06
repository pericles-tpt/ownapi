#!/bin/sh

# 1. Build frontend code, TS, SCSS
if [ ! -d "./node_modules" ]; then
    npm i
fi
npx webpack && cd ..

# 2. Set environment to dev
export IS_DEV=true