FROM golang

RUN go get -u github.com/golang/dep/cmd/dep


ADD . /go/src/github.com/duhruh/cluster

WORKDIR /go/src/github.com/duhruh/cluster


CMD ["go", "run", "main.go"]

