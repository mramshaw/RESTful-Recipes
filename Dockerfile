FROM golang:1.14.3

RUN go get golang.org/x/lint/golint

RUN go get github.com/julienschmidt/httprouter
RUN go get github.com/lib/pq

EXPOSE 8080
