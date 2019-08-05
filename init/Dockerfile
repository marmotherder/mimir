FROM alpine:latest

RUN mkdir /etc/certs

RUN wget -O /usr/bin/cfssl https://pkg.cfssl.org/R1.2/cfssl_linux-amd64 && chmod 777 /usr/bin/cfssl
RUN wget -O /usr/bin/cfssljson https://pkg.cfssl.org/R1.2/cfssljson_linux-amd64 && chmod 777 /usr/bin/cfssljson
RUN wget -O /usr/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/$(wget -q -O - https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl && chmod 777 /usr/bin/kubectl
RUN apk update && apk add gettext jq && rm -rf /var/cache/apk/*

COPY setup.sh /etc/certs/setup.sh
RUN chmod 700 /etc/certs/setup.sh

WORKDIR /etc/certs

CMD ["/etc/certs/setup.sh"]