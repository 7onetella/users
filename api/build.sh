#!/bin/sh
set +x

rm -f users
rm -f authp

cd ../accounts
./build.sh
cd ../api

echo copying the emberjs asset from ui/dist/*
mkdir -p accounts/
rm -r accounts/*
cp -r ../accounts/dist/* ./accounts/

# Jenkins path can be missing this
PATH=$PATH:.:~/bin

echo generate binddata.go file
go-bindata-assetfs accounts/...

# get gox for cross compilation
go get -u github.com/mitchellh/gox

echo cross compile
"${GOPATH}"/bin/gox -osarch="linux/amd64"

mv api_linux_amd64 authp
cp authp ../../box/playbooks/roles/authp/files/authp
scp -i ~/.aws/7onetella.pem authp ubuntu@authp.7onetella.net:/root/


