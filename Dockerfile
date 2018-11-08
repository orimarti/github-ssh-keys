FROM golang:1.11.1 as builder
WORKDIR /go/src/app
COPY main.go .

RUN go get -d ./... \
    && env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /go/bin/app ./...

FROM scratch
COPY --from=builder /go/bin/app /go/bin/app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/go/bin/app"]  
