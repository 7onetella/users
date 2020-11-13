#!/bin/bash

echo rebuilding documentation
cp redoc/redoc.html accounts/
swagger generate spec -i input.yml -m > accounts/swagger.json
go-bindata-assetfs accounts/...

refresh run