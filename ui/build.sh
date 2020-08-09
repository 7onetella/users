#!/bin/sh
set +x

npm install
rm -rf dist/*
ember build --environment=production 
