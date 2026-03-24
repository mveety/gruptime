#!/bin/sh

go build -gcflags="-N -l" -ldflags="-s"

