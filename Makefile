build:
	cd cmd/andrew && go build

run:
	go run ./cmd/andrew/main.go $$HOME/Developer/playtechnique/website/content
