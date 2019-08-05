#!/usr/bin/env sh

envsubst < templates/csr.json > csr.json

cfssl genkey csr.json | cfssljson -bare server

export SERVER_CERT=$(cat server.csr | base64 | tr -d '\n')

envsubst < templates/csr.yaml > csr.yaml

kubectl apply -f csr.yaml

kubectl certificate approve "${RELEASE}"

export CA_BUNDLE=$(cat /var/run/secrets/kubernetes.io/serviceaccount/ca.crt | base64 | tr -d '\n')
envsubst < templates/webhook.yaml > webhook.yaml

kubectl apply -f webhook.yaml

kubectl get csr "${RELEASE}" -o jsonpath='{.status.certificate}' | base64 -d > server-cert.pem

cp server-key.pem output/server-key.pem
cp server-cert.pem output/server-cert.pem
