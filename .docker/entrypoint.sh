#!/bin/bash

until PGHOST=${HOST_DB} PGDATABASE=${POSTGRES_DB} PGPORT=${PORT_DB} PGUSER=${POSTGRES_USER} PGPASSWORD=${POSTGRES_PASSWORD} psql -c 'SELECT 1' >> /dev/null; do sleep 5; done;

migrate -source file://${PWD}/migrations/ -database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${HOST_DB}:${PORT_DB}/${POSTGRES_DB}?sslmode=disable up

cleanup_on_sigterm() {
    if [ -n "$PROGRAM_PID" ]; then
        kill "$PROGRAM_PID"
        wait "$PROGRAM_PID"
    fi
    exit 0
}

trap cleanup_on_sigterm SIGTERM

./ew &
PROGRAM_PID=$!

wait "$PROGRAM_PID"