#!/bin/sh

kubectl patch jsonserver app-my-server \
  --type=json \
  --patch-file patches/add_person.json

sleep 30

kubectl patch jsonserver app-my-server \
  --type=json \
  --patch-file patches/increase_replicas.json