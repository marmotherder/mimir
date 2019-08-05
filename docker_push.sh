#!/bin/bash
echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
docker push marmotherder/mimir-init:latest
docker push marmotherder/mimir:latest