#!/bin/sh
set +x

# delete the go binary
rm -f api_linux_amd64

# build emberjs app
cd ../accounts
./build.sh

# copy emberjs app deployment package to api/accounts
cd ../api
echo copying the emberjs asset from ui/dist/*
mkdir -p accounts/
rm -r accounts/*
cp -r ../accounts/dist/* ./accounts/

# Jenkins path can be missing this
PATH=$PATH:.:~/bin

echo generate binddata.go file
swagger generate spec -m > accounts/swagger.json
cp redoc/redoc.html accounts
go-bindata-assetfs accounts/...

# get gox for cross compilation
go get -u github.com/mitchellh/gox

echo cross compile
"${GOPATH}"/bin/gox -osarch="linux/amd64"

mv api_linux_amd64 api_linux_amd64
cp api_linux_amd64 ../../box/playbooks/roles/authp/files/api_linux_amd64
scp -i ~/.aws/7onetella.pem api_linux_amd64 ubuntu@authp.7onetella.net:/root/


