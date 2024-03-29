############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder
# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/package/app/
COPY . .
# Fetch dependencies.
# Using go get.
RUN go get -d -v
# Build the binary.
COPY cmd/showcash-api/migrations /go/bin/migrations/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w -s' -o /go/bin/showcash-api ./cmd/...
############################
# STEP 2 build a small image
############################
FROM scratch
# Copy our static executable.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/migrations/ /go/bin/migrations/
COPY --from=builder /go/bin/showcash-api /go/bin/showcash-api
# Run the showcash-api binary.
ENTRYPOINT ["/go/bin/showcash-api"]
