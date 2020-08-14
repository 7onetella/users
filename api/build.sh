#!/bin/sh
set +x

echo copying the emberjs asset from ui/dist/*
mkdir -p ui/
cp -r ../ui/dist/* ./ui/

# Jenkins path can be missing this
PATH=$PATH:.:~/bin

echo generate binddata.go file
go-bindata-assetfs ui/...

go build -o users

