#!/bin/bash

REPO=$(aws ecr describe-repositories | jq -r '.repositories[].repositoryUri')

docker build -t $REPO:latest --platform=linux/amd64 .
docker push $REPO:latest

aws lambda update-function-code \
    --function-name  go-scrape-lambda \
    --image-uri $REPO:latest

