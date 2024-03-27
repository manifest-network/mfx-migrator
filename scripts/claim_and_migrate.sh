#!/usr/bin/env bash

WORKDIR=/jobs

cd "$WORKDIR" || exit 1

# Claim some work from queue
mfx-migrator claim

# Run the migration for each JSON file in the workdir
find . -name '*.json' ! -name "config.json" -print0 | xargs -0 -I{} -P 1 bash -c 'mfx-migrator migrate --uuid $(basename "{}" .json)'
