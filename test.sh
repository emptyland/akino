#!/bin/bash

cd ../../../..
GOPATH=`pwd` go test -v ./...
