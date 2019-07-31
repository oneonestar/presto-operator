#!/usr/bin/env bash
kubectl delete configmap coordinator-config
kubectl delete configmap worker-config
kubectl delete configmap catalog

kubectl create configmap coordinator-config --from-file=coordinator_config
kubectl create configmap worker-config --from-file=worker_config
kubectl create configmap catalog --from-file=catalog