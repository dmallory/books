FROM golang:1.8-alpine
ADD . /go/src/github.com/dmallory/books
RUN apk add --no-cache git
RUN go get github.com/gorilla/mux gopkg.in/mgo.v2/bson github.com/BurntSushi/toml
RUN go install github.com/dmallory/books

FROM alpine:latest
COPY --from=0 /go/bin/books .
EXPOSE 3000
CMD ["./books"]

FROM mvertes/alpine-mongo
VOLUME /var/lib/mongodb
CMD ["mongod","--dbpath=/var/lib/mongodb"]
