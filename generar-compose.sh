#!/bin/bash

if [ "$#" -ne 2 ]; then
    echo "Uso: $0 <nombre_archivo_salida> <cantidad_clientes>"
    exit 1
fi

OUTPUT_FILE=$1
NUM_CLIENTS=$2

cat > "$OUTPUT_FILE" << EOF
name: tp0
services:
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
      - LOGGING_LEVEL=DEBUG
    networks:
      - testing_net

EOF

for ((i=1; i<=$NUM_CLIENTS; i++)); do
    cat >> "$OUTPUT_FILE" << EOF
  client$i:
    container_name: client$i
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID=$i
      - CLI_LOG_LEVEL=DEBUG
    networks:
      - testing_net
    depends_on:
      - server

EOF
done

cat >> "$OUTPUT_FILE" << EOF
networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
EOF
