#!/usr/bin/env bash

APP=mfx-migrator
WORKDIR=/jobs

cd "$WORKDIR" || exit 1

# Claim some work
$APP claim

# Run the migration
find . -name '*.json' ! -name "config.json" -print0 | xargs -0 -I{} -P 1 bash -c '$APP migrate --uuid $(basename "{}" .json)'
