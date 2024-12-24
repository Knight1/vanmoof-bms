#!/bin/zsh

set -e #exit on error
#set -x

GV=$(git tag || echo 'N/A')
if [[ $GV =~ [^[:space:]]+ ]];
then
    GitTag=${BASH_REMATCH[0]}
fi

GH=$(git log -1 --pretty=format:%h || echo 'N/A')
if [[ GH =~ 'fatal' ]];
then
    CommitHash=N/A
else
    CommitHash=$GH
fi

GoVersion=N/A
if [[ $(go version) =~ [0-9]+\.[0-9]+\.[0-9]+ ]];
then
    GoVersion=${BASH_REMATCH[0]}
fi

export GOOS=linux
export GOARCH=arm64
#export GOARM=8
export VERSION=$(git rev-parse --short HEAD)
BUILDTIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
TRG_PKG='main'

FLAG="-X $TRG_PKG.BuildTime=$BUILDTIME"
FLAG="$FLAG -X $TRG_PKG.CommitHash=$CommitHash"
FLAG="$FLAG -X $TRG_PKG.GoVersion=$GoVersion"
FLAG="$FLAG -X $TRG_PKG.GOOS=$GOOS"
FLAG="$FLAG -X $TRG_PKG.GOARCH=$GOARCH"
FLAG="$FLAG -s -w"



garble build -ldflags ${FLAG}
#go test -v ./...
#upx -9 --best bms
scp bms root@10.0.0.7:/root/VanMoooof-bms/bms
