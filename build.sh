#!/bin/zsh

set -e #exit on error
set -x

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

export GOOS=linux
export GOARCH=arm
export GOARM='6,hardfloat'
export VERSION=$(git rev-parse --short HEAD)
BUILDTIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
TRG_PKG='main'

FLAG="-X $TRG_PKG.BuildTime=$BUILDTIME"
FLAG="$FLAG -X $TRG_PKG.CommitHash=$CommitHash"
FLAG="$FLAG -X $TRG_PKG.GitTag=$GitTag"
FLAG="$FLAG -X $TRG_PKG.GOOS=$GOOS"
FLAG="$FLAG -X $TRG_PKG.GOARCH=$GOARCH"
FLAG="$FLAG -s -w"



go build -v -ldflags ${FLAG}
scp bms knight@192.168.8.145:/home/knight/bms
