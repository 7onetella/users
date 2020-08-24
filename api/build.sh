#!/bin/sh
set +x

rm -f users
rm -f authp

cd ../ui
./build.sh
cd ../api

echo copying the emberjs asset from ui/dist/*
mkdir -p ui/
rm -r ui/*
cp -r ../ui/dist/* ./ui/

# Jenkins path can be missing this
PATH=$PATH:.:~/bin

echo generate binddata.go file
go-bindata-assetfs ui/...

# get gox for cross compilation
go get -u github.com/mitchellh/gox

echo cross compile
"${GOPATH}"/bin/gox -osarch="linux/amd64"

mv api_linux_amd64 authp
cp authp ../../box/playbooks/roles/authp/files/authp
scp -i ~/.aws/7onetella.pem authp ubuntu@authp.7onetella.net:/root/


