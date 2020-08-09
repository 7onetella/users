#!/bin/bash

go build -gcflags "all=-N -l" github.com/7onetella/users/api

dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec ./api

