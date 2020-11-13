# syntax = docker/dockerfile:experimental

FROM golang:alpine AS builder
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git build-base

# Create appuser.
ENV USER=appuser
ENV UID=10001 
RUN adduser \    
    --disabled-password \    
    --gecos "" \    
    --home "/nonexistent" \    
    --shell "/sbin/nologin" \    
    --no-create-home \    
    --uid "${UID}" \    
    "${USER}"

WORKDIR $GOPATH/src/chronicler/worker/

COPY go.mod ./
RUN go mod download
RUN go mod verify
RUN go get -u -v github.com/mattn/go-sqlite3

COPY cmd cmd
COPY pkg pkg
COPY internal internal 

# Build the binary.
RUN --mount=type=cache,uid=10001,target=/go/.cache/go-build \
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/worker ./cmd/worker

FROM python:alpine

RUN apk update && apk add --no-cache \
        ca-certificates \
        ffmpeg \
        && pip3 install -U youtube-dl \
        && \ 
    rm -rf /var/cache/apk/*

# Import the user and group files from the builder.
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
# Copy our executable.
COPY --from=builder /go/bin/worker /go/bin/worker

# Use an unprivileged user.
USER appuser:appuser
WORKDIR /workdir
ENTRYPOINT ["/go/bin/worker"]