#!/bin/sh
rm -rf _config/ projected _frontend/

bsdtar -xf build.zip  -s'|[^/]*/ [^_/]*/||'