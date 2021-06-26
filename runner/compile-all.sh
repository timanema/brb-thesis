#!/bin/bash

CGO_ENABLED=0 go build .
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build .
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o rp-runner-mac .
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o rp-runner-mac-m1 .
