#!/bin/sh -e

set -x

target_env=${1:-production}

npm install
rm -rf dist/*
ember build --environment=${target_env} 
