#!/bin/sh

kubectl patch jsonserver app-my-server \
  --type=json \
  --patch-file add_person.json