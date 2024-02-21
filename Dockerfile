FROM golang:1.21 AS base

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum* ./
RUN go mod download && go mod verify
COPY . .

RUN CGO_ENABLED=0 go build -v -o /usr/local/bin/andrew ./cmd/andrew/main.go

FROM scratch
COPY --from=base /etc/passwd /etc/passwd
USER 1000
COPY --from=base /usr/local/bin/andrew /usr/local/bin/andrew

#By default, andrew will serve the pwd.
ENTRYPOINT ["/usr/local/bin/andrew"]
