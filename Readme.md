## Code generation
```/home/saima/go/src/k8s.io/code-generator/generate-groups.sh deepcopy,client,informer,lister github.com/saima-s/data-restore/pkg/client  github.com/saima-s/data-restore/pkg/apis saima.dev.com:v1 --go-header-file /home/saima/go/src/k8s.io/code-generator/examples/hack/boilerplate.go.txt```


## Controller-gen
```controller-gen paths=github.com/saima-s/data-restore/pkg/apis/saima.dev.com/v1  crd:crdVersions=v1 output:crd:artifacts:config=manifests```

## Commands to run the application
1. go build
2. ./data-restore


## Docker commands

1. docker build -t data-restore:6.0.0 .
2. docker tag  data-restore:6.0.0 sultanasaima/data-restore:6.0.0
3. docker push sultanasaima/data-restore:6.0.0



