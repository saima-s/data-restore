FROM golang:latest

# ENV GO111MODULE=off



# ARG repo="${GOPATH}/src/github.com/saima-s/data-restore"

# RUN mkdir -p $repo

# WORKDIR $GOPATH/src/k8s.io/code-generator

# VOLUME $repo
WORKDIR /home

RUN go get k8s.io/code-generator; exit 0
RUN go get k8s.io/apimachinery; exit 0

COPY . /home

RUN go build -o data-restore

CMD [ "./data-restore" ]
