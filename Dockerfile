FROM golang:1.10.1-alpine

RUN apk add --no-cache --update git make

RUN go get golang.org/x/lint/golint

RUN go get github.com/julienschmidt/httprouter
RUN go get github.com/lib/pq

EXPOSE 8080
