# data-restore

## Introduction

The data-restore is a custom kubernetes controller written in Go.
Following are the key features of this controller:
1. Watch on newly create CR(Snap, Restore).
2. On Addition of custom(Snap) CR, it takes snapshot of PVC.
3. On Addition of custom(restore) CR, it takes the backup the snapshot created into a new PVC.

## Run code

### Pre-requisite
 1. Install minikube locally: https://www.linuxbuzz.com/install-minikube-on-ubuntu/

### Locally

 To run the data-restore controller on local machine:
 1. Open a terminal.
 2. ```bash
       go build```
 3. ```bash
       ./data-restore```
 4. Create a snapCR custom resource by following command:
    ```bash
        cd manifests/kubectl create -f snapCR.yaml```
 5. Create a restoreCR custom resource by following command:
    ```bash
        cd manifests/kubectl create -f restoreCR.yaml```

 ### Cluster

 To run application on cluster:
 1. Dockerize application by writing Dockerfile.
 2. Create service account, clusterrole and clusterrolebinding to access custom resource and watch it by running following command:
    ```bash
        cd manifests/kubectl create -f sa.yaml```
    ```bash
        cd manifests/kubectl create -f role.yaml```
 3. Deploy the application by running following command:
    ```bash
       cd manifests/kubectl create -f deployment.yaml```
 4. exec into pod and run:
    ```bash
        ./data-restore``` 
 
 



















## Code generation
```/home/saima/go/src/k8s.io/code-generator/generate-groups.sh deepcopy,client,informer,lister github.com/saima-s/data-restore/pkg/client  github.com/saima-s/data-restore/pkg/apis saima.dev.com:v1 --go-header-file /home/saima/go/src/k8s.io/code-generator/examples/hack/boilerplate.go.txt```


## Controller-gen
```controller-gen paths=github.com/saima-s/data-restore/pkg/apis/saima.dev.com/v1  crd:crdVersions=v1 output:crd:artifacts:config=manifests```

## Commands to run the application
1. go build
2. ./data-restore


## Docker commands

1. docker build -t data-restore:10.0.0 .
2. docker tag  data-restore:10.0.0 sultanasaima/data-restore:10.0.0
3. docker push sultanasaima/data-restore:10.0.0







