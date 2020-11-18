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

WORKDIR $GOPATH/src/github.com/mike-dunton/
COPY go.mod ./

RUN go mod download
RUN go mod verify
RUN go get -u -v github.com/mattn/go-sqlite3

COPY cmd cmd
COPY pkg pkg 
COPY internal internal 

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/web ./cmd/web

FROM node:14.1-alpine AS node-builder

WORKDIR /opt/web
COPY ./web/package.json ./web/package-lock.json ./
RUN npm install

ENV PATH="./node_modules/.bin:$PATH"

COPY ./web/ ./
RUN npm run build

FROM alpine

# Import the user and group files from the builder.
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
# Copy our static executable.
COPY --from=builder /go/bin/web /go/bin/web
COPY --from=node-builder /opt/web/build /usr/share/html

RUN mkdir /downloads && chown -R appuser /downloads
RUN mkdir /data  && chown -R appuser /data

# Use an unprivileged user.
USER appuser:appuser
WORKDIR /workdir
ENTRYPOINT ["/go/bin/web"]