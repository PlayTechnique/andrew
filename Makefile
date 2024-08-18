build:
	cd cmd/andrew && go build

run:
	./cmd/andrew/andrew $$HOME/Developer/playtechnique/website/website

test:
	docker run --rm -v "$${PWD}":/usr/src/andrew -w /usr/src/andrew golang:1.23 go test

SSL_DIR := ./test-ssl-cert
KEY_FILE := server.key
CERT_FILE := server.crt

generate-ssl-cert:
	@mkdir -p $(SSL_DIR)
	@cd $(SSL_DIR) && \
	openssl genrsa -out $(KEY_FILE) 2048 && \
	openssl ecparam -genkey -name secp384r1 -out $(KEY_FILE) && \
	openssl req -new -x509 -sha256 -key $(KEY_FILE) -out $(CERT_FILE) -days 3650

clean-ssl-cert:
	@rm -rf $(SSL_DIR)