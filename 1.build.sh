#!/bin/bash

mkdir -p out

docker compose build 
docker compose up -d

./9.clean.sh
