FROM golang:1.15.4

RUN go get golang.org/x/lint/golint

RUN go get github.com/jmoiron/sqlx
RUN go get github.com/julienschmidt/httprouter
RUN go get github.com/lib/pq
RUN go get github.com/stretchr/testify/assert

EXPOSE 8080
