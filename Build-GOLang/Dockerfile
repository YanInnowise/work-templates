FROM golang:1.19.1-alpine AS builder
ENV GO111MODULE=auto
WORKDIR /go/src/

RUN apk add --no-cache git 
COPY main.go ./
 
#init is initializing and writting new go.mod in current dir.
RUN go mod init main.go

RUN go get -d -v github.com/caarlos0/env \
    && go get -d -v github.com/prometheus/client_golang/prometheus \
    && go get -d -v github.com/prometheus/client_golang/prometheus/promhttp \
    && go get -d -v github.com/lib/pq \   
    && go get -d -v database/sql \
    && go get -d -v fmt \
    && go get -d -v log \
    && go get -d -v net/http \
    && go get -d -v time
 	
RUN go build -o app .

FROM alpine:3.10.3
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/app ./
CMD ["./app"]
  
