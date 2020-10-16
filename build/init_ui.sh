#!/bin/sh

npm run build && \
npm install -g http-server && \
http-server -p 80 /app/ui/dist
