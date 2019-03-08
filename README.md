# presto-controller
This is a prototype of Presto K8S Operator. Tested and developed in GKE.
This project follows the approach used in `kubernetes/sample-controller`.

# How to start
## Start the presto-controller
Compile the presto-controller
```bash
$ dep ensure  # This project is using dep for dependency management
$ go build -o presto-controller .
```

Start the presto-controller
```bash
$ ./presto-controller -kubeconfig=<PATH_TO_YOUR_KUBECONFIG> -logtostderr -v 4
```

## Create a Presto CRD
Create a CRD Definition in your K8S cluster.
```bash
kubectl create -f artifacts/crd.yaml
```

Create a Presto resource
```bash
kubectl create -f artifacts/example1.yaml
```

## Example
```bash
# Create the configmaps for presto
$ kubectl create configmap coordinator-config --from-file=coordinator_config
$ kubectl create configmap worker-config --from-file=worker_config
$ kubectl create configmap catalog --from-file=catalog

# Specify the configmaps in CRD
$ cat artifacts/example1.yaml
apiVersion: prestocontroller.prestosql.io/v1alpha1
kind: Presto
metadata:
  name: presto-cluster
spec:
  clusterName: star-presto
  image: "gcr.io/learn-227713/presto:latest"
  replicas: 2
  coordinatorConfig: coordinator-config   ## Config file for coordinator
  workerConfig: worker-config             ## Config file for worker
  catalogConfig: catalog                  ## Config file for catalog

$ kubectl create -f artifacts/example1.yaml
presto.prestocontroller.prestosql.io/presto-cluster created

$ kubectl get presto
NAME             DESIRED   CURRENT
presto-cluster   2         2

$ kubectl get rs
NAME                            DESIRED   CURRENT   READY   AGE
star-presto-coordinator-r9jtl   1         1         0       6s
star-presto-worker-9vv65        2         2         0       5s

$ kubectl get pods
NAME                                  READY   STATUS    RESTARTS   AGE
star-presto-coordinator-r9jtl-sbwjt   1/1     Running   0          22s
star-presto-worker-9vv65-7ddmq        1/1     Running   0          22s
star-presto-worker-9vv65-kcb27        1/1     Running   0          22s
```

# WIP
## Deploy operator using pre-build docker image
Setup the service account if you are using RBAC. Then create a presto-operator deployment. You may refer to `artifacts/operator.yaml`

# Memo
The Dockerfile in `docker/presto/` builds the Docker image of Presto being used by this project. The Dockerfile in `docker/` builds the Docker image of Presto-operator.

```text
Presto Resource = 
    Presto Coordinator ReplicaSet
    + Presto Worker ReplicaSet
    + Presto Coordinator Service

Presto Coordinator ReplicaSet =
    Presto docker image
    + Coordinator ConfigMap/Secret

Presto Worker ReplicaSet = 
    Presto docker image
    + Worker ConfigMap/Secret
```
