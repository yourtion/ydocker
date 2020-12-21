#!/bin/sh

set -e

go clean -testcache

go test -cover $(go list ./... | grep -v /demo/)