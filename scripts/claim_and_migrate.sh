#!/usr/bin/env bash

WORKDIR=/jobs
QUARANTINE_DIR=/quarantine

cd "$WORKDIR" || exit 1

# Claim some work from queue
mfx-migrator claim

# Run the migration for each JSON file in the workdir
find . -name '*.json' ! -name "config.json" -print0 | xargs -0 -I{} -P 1 bash -c 'mfx-migrator migrate --uuid $(basename "{}" .json)'

# Check each JSON file for a failed status and a non-empty error
for file in *.json; do
    [[ -e "$f" ]] || break # Exit if no files found

    if [ "$file" == "config.json" ]; then
        continue
    fi

    status=$(jq -r '.status' "$file")
    error=$(jq -r '.error' "$file")

    if [[ "$status" -eq 5 ]] && [[ -n "$error" ]]; then
        # If the quarantine directory doesn't exist, create it
        mkdir -p "$QUARANTINE_DIR"

        # Move the file to the quarantine directory
        mv "$file" "$QUARANTINE_DIR/"
    fi
done
