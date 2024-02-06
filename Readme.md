# data-restore

## Table of Contents
[Introduction](Introduction)

[Getting Started](Gettingstarted)

[Run Code](Runcode)

[Reference](References)

## Introduction

The data-restore is a custom kubernetes controller written in Go.
Following are the key features of this controller:
1. Watch on newly create CR(Snap, Restore).
2. On Addition of custom(Snap) CR, it takes snapshot of PVC.
3. On Addition of custom(restore) CR, it takes the backup the snapshot created into a new PVC.

# Getting started
Below instruction will help in setting up project locally as well as up and running in cluster:

## Dependencies:
To run the controller on Local System,need to install following Software Dependencies.

[Go](https://go.dev/doc/install)

[Docker](https://docs.docker.com/engine/install/ubuntu/)

[Minikube](https://www.linuxbuzz.com/install-minikube-on-ubuntu/)


# Run code

## Locally

 To run the data-restore controller on local machine:
 1. Open a terminal and run below commands: 
   ```sh
   go build
    
   ```
   ```sh
   ./data-restore
    
   ```
 2. Create a snapCR custom resource by following command:

   ```sh
    cd manifests/kubectl create -f snapCR.yaml
   ```
 3. Create a restoreCR custom resource by following command:

   ```sh
    cd manifests/kubectl create -f restoreCR.yaml
   ```

## Cluster

 Start cluster and enable addon: 
 
  ```sh
   minikube start
  ```


  ```sh
  minikube addons enable volumesnapshots
  ```


  ```sh
  minikube addons enable csi-hostpath-driver
  ```

 To run application on cluster:
 1. Dockerize application by writing Dockerfile. Build and push image to docker hub repository [here](#Docker-commands).
 2. Create service account, clusterrole and clusterrolebinding to access custom resource and watch it by running following command:

  ```sh
  cd manifests/kubectl create -f sa.yaml
  ```
    
   ```sh
   cd manifests/kubectl create -f role.yaml
   ```
 4. Deploy the application by running following command:
    
   ```sh
    cd manifests/kubectl create -f deployment.yaml
   ```
 6. exec into pod and run:
    
   ```sh
   ./data-restore
   ``` 

# References
1. https://pkg.go.dev/k8s.io/client-go
2. https://pkg.go.dev/k8s.io/apimachinery
3. https://pkg.go.dev/github.com/mitchellh/go-homedir
4. https://minikube.sigs.k8s.io/docs/tutorials/volume_snapshots_and_csi/
5. https://pkg.go.dev/github.com/kubernetes-csi/external-snapshotter/v6


## Docker commands

1. docker build -t data-restore:10.0.0 .
2. docker tag  data-restore:10.0.0 sultanasaima/data-restore:10.0.0
3. docker push sultanasaima/data-restore:10.0.0







