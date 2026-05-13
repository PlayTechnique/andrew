test name="":
	#!/usr/bin/env sh
	if [ -z "{{name}}" ]; then
		go test -v ./...
	else
		go test -v -run "{{name}}" ./...
	fi
