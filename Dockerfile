FROM alpine:latest

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

COPY ${GOPATH}/bin/mimir /bin/mimir
RUN chmod 700 /bin/mimir

ENTRYPOINT ["/bin/mimir"]