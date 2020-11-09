FROM golang:alpine AS builder
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

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
COPY ./worker .

RUN go mod download
RUN go mod verify

# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/chronicler-worker

FROM debian:stable-slim

ENV LC_ALL=C.UTF-8 \
    LANG=C.UTF-8 \
    LANGUAGE=en_US:en

RUN apt-get update -y && \
    apt-get install --no-install-recommends -y \
        ca-certificates \
        ffmpeg \
        python3 \
        python3-pip \
        && pip3 install -U youtube-dl \
        && \ 
    apt-get autoremove -y && \
    apt-get clean -y && \
    rm -rf /var/lib/apt/lists/*

# Import the user and group files from the builder.
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
# Copy our static executable.
COPY --from=builder /go/bin/chronicler-worker /go/bin/chronicler-worker

# Use an unprivileged user.
USER appuser:appuser
WORKDIR /workdir
ENTRYPOINT ["/go/bin/chronicler-worker"]