#!/bin/bash

set -e
set -v

export ENV="local"

# run service
go run . --config=./config/local.json
