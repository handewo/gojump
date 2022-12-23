#!/bin/bash

rm -rf gojumpdb

go build -trimpath -ldflags="-s -w" cmd/gojump.go

go build -trimpath -ldflags="-s -w" cmd/initDB/initDB.go

./initDB
