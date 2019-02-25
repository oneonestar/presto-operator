# presto-controller

```text
Presto Resource = 
    Presto Coordinator ReplicaSet (ConfigMap/Secret)
    + Presto Coordinator Service
    + Presto Worker ReplicaSet (ConfigMap/Secret)

Presto Coordinator ReplicaSet =
    Presto docker image
    + Coordinator ConfigMap/Secret

Presto Worker ReplicaSet = 
    Presto docker image
    + Worker ConfigMap/Secret

```
