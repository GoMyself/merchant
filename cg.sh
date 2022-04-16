#! /bin/bash

#git checkout main
#git pull origin main
#git submodule init
#git submodule update --remote

PROJECT="merchant"
GitReversion=`git rev-parse HEAD`
BuildTime=`date +'%Y.%m.%d.%H%M%S'`
BuildGoVersion=`go version`

go build -ldflags "-X main.gitReversion=${GitReversion}  -X 'main.buildTime=${BuildTime}' -X 'main.buildGoVersion=${BuildGoVersion}'" -o $PROJECT

scp -i /home/gocloud-yiy-rich $PROJECT p3test@34.92.240.177:/home/centos/workspace/cg/merchant/merchant_cg
ssh -i /home/gocloud-yiy-rich p3test@34.92.240.177 "sh /home/centos/workspace/cg/merchant/cg.sh"