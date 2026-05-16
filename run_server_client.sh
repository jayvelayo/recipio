#!/bin/bash

function start_server() {
    cd 'backend';
    go run cmd/recipio-server/recipio_server.go
}

function start_client() {
    cd 'frontend'
    npm run dev
}

echo "Running server and client.."
start_server &
start_client
