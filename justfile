test name="":
	#!/usr/bin/env sh
	if [ -z "{{name}}" ]; then
		go test ./...
	else
		go test -v -run "{{name}}" ./...
	fi

serve:
	go run ./cmd/andrew ../playtechnique/website/content
