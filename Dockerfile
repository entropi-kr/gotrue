FROM golang:1.22-alpine as build
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux

RUN apk add --no-cache make git

WORKDIR /go/src/github.com/entropi-kr/gotrue

# Pulling dependencies
COPY ./Makefile ./go.* ./
RUN make deps

# Building stuff
COPY . /go/src/github.com/entropi-kr/gotrue
RUN make build

FROM alpine:3.7
RUN adduser -D -u 1000 entropi

RUN apk add --no-cache ca-certificates
COPY --from=build /go/src/github.com/entropi-kr/gotrue/gotrue /usr/local/bin/gotrue
COPY --from=build /go/src/github.com/entropi-kr/gotrue/migrations /usr/local/etc/gotrue/migrations/

ENV GOTRUE_DB_MIGRATIONS_PATH /usr/local/etc/gotrue/migrations

USER entropi
CMD ["gotrue"]
