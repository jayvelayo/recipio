#!/usr/bin/env sh
# Start crond in the background, then hand off to the server.
crond -l 8
exec ./recipio-server
