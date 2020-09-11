#!/bin/bash

swagger generate spec -m > accounts/swagger.json
go-bindata-assetfs accounts/...
