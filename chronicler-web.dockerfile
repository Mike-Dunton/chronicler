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

WORKDIR /app
COPY go.mod ./

RUN go mod download
RUN go mod verify
RUN go get -u -v github.com/mattn/go-sqlite3

COPY web .
# Build the binary.
# RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -ldflags '-w -linkmode external -extldflags "-static"' -o /go/bin/web

RUN --mount=type=cache,uid=10001,target=/go/.cache/go-build \
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/web

FROM node:14.1-alpine AS node-builder

WORKDIR /opt/web
COPY ./webui/package.json ./webui/package-lock.json ./
RUN npm install

ENV PATH="./node_modules/.bin:$PATH"

COPY ./webui/ ./
RUN npm run build

FROM alpine

# Import the user and group files from the builder.
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
# Copy our static executable.
COPY --from=builder /go/bin/web /go/bin/web
COPY --from=node-builder /opt/web/build /usr/share/html

COPY ./webui /src

# Use an unprivileged user.
USER appuser:appuser
WORKDIR /workdir
ENTRYPOINT ["/go/bin/web"]