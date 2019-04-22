FROM alpine:latest

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

COPY /tmp/mimir /bin/mimir
RUN chmod 700 /bin/mimir

ENTRYPOINT ["/bin/mimir"]