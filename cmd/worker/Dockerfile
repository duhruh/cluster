FROM golang:alpine as dev

RUN apk add git

RUN go get -u github.com/golang/dep/cmd/dep

ADD . /go/src/github.com/duhruh/cluster

WORKDIR /go/src/github.com/duhruh/cluster

RUN go build -i -o worker cmd/worker/main.go

FROM alpine

RUN apk --no-cache \ 
        add ca-certificates

COPY --from=dev /go/src/github.com/duhruh/cluster/worker /app/worker

CMD ["/app/worker"]
