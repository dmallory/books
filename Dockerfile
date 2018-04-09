FROM golang:1.10.1-alpine
ADD . /go/src/github.com/dmallory/books
RUN apk add --no-cache git mongodb
RUN go get github.com/gorilla/mux gopkg.in/mgo.v2/bson github.com/BurntSushi/toml
RUN go install github.com/dmallory/books
EXPOSE 3000
VOLUME /var/lib/mongodb
CMD ["/go/src/github.com/dmallory/books/start"]
