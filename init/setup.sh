#!/usr/bin/env sh

envsubst < templates/csr.json > csr.json

if [ -f "./templates/ca-csr.json" ]; then
    envsubst < templates/ca-csr.json > ca-csr.json
    envsubst < templates/ca-config.json > ca-config.json
    cfssl gencert -initca ca-csr.json | cfssljson -bare ca
    cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json csr.json | cfssljson -bare server
    export CA_BUNDLE=$(cat ca.pem | base64 | tr -d '\n')
    mv server.pem server-cert.pem
else
    cfssl genkey csr.json | cfssljson -bare server
    export SERVER_CERT=$(cat server.csr | base64 | tr -d '\n')
    envsubst < templates/csr.yaml > csr.yaml
    kubectl apply -f csr.yaml
    kubectl certificate approve "${RELEASE}"
    kubectl get csr "${RELEASE}" -o jsonpath='{.status.certificate}' | base64 -d > server-cert.pem
    export CA_BUNDLE=$(cat /var/run/secrets/kubernetes.io/serviceaccount/ca.crt | base64 | tr -d '\n')
fi

envsubst < templates/webhook.yaml > webhook.yaml
kubectl apply -f webhook.yaml

cp server-key.pem output/server-key.pem
cp server-cert.pem output/server-cert.pem