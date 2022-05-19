FROM golang:1.17.3-alpine as build

COPY . /usr/src/code
WORKDIR /usr/src/code
RUN apk add build-base
RUN go build ./cmd/token_service/...

FROM alpine:latest as production-build

RUN apk add --update --no-cache supervisor && rm -rf /var/cache/apk/*

RUN mkdir /opt/code
COPY --from=build /usr/src/code/token_service /opt/code/token_service
ADD supervisord.conf /etc/supervisord.conf

# This command runs your application, comment out this line to compile only
CMD ["/usr/bin/supervisord","-n", "-c", "/etc/supervisord.conf"]

LABEL Name=pengine Version=0.0.1