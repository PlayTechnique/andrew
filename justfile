test name="":
    #!/usr/bin/env sh
    if [ -z "{{ name }}" ]; then
    	go test ./...
    else
    	go test -v -run "{{ name }}" ./...
    fi

serve:
    go run ./cmd/andrew ../playtechnique/website/content

run:
    #!/usr/bin/env sh
    go run ./cmd/andrew ../playtechnique/website/content &
    server_pid=$!
    trap 'kill $server_pid 2>/dev/null' EXIT INT TERM
    until curl -sf http://localhost:8080 >/dev/null 2>&1; do
    	if ! kill -0 $server_pid 2>/dev/null; then
    		echo "server exited before it came up" >&2
    		exit 1
    	fi
    	sleep 0.2
    done
    open http://localhost:8080
    wait $server_pid

rssServe:
    go run ./cmd/andrew --rssdir blog ../playtechnique/website/content