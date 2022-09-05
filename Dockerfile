# Build stage
FROM golang:1.18 as builder
LABEL maintainer="Sergey Prokhorov <lisforlinux@gmail.com>"

WORKDIR /go/src/github.com/sprokhorov/helmctl
COPY . ./

RUN GOARCH=amd64 CGO_ENABLED=0 GOOS=linux go build -o helmctl main.go


# Main stage
FROM spr0khorov/ci-base:v0.0.1
LABEL maintainer="Sergey Prokhorov <lisforlinux@gmail.com>"

COPY --from=builder /go/src/github.com/sprokhorov/helmctl/helmctl /usr/local/bin/
