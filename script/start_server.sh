#!/bin/bash

set -e
set -v

export ENV="local"

# run service
nohup go run .. --config=../config/server.json > ../log/out &
