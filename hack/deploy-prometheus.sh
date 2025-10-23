#!/bin/sh

sudo kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/bundle.yaml

sudo kubectl apply -f config/prometheus/prometheus-instance.yaml

# To access the frontend on the host's network, run:
# kubectl port-forward svc/prometheus-operated 9090:9090
