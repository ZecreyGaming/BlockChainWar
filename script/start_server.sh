#!/bin/bash

set -e
set -v

# run service
nohup go run .. --config=../config/server.json > ../log/out &
