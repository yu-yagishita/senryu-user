FROM golang:1.14.2-alpine3.11
ENV sourcesdir /go/src/github.com/yu-yagishita/senryu-user
ENV MONGO_HOST user-db:27017
ENV HATEAOS user
ENV USER_DATABASE mongodb
ENV GO111MODULE=on

WORKDIR ${sourcesdir}
COPY . .

RUN apk add --no-cache alpine-sdk
RUN apk update
RUN apk add git

# Golang ホットリロード(freshのインストール)
RUN go get github.com/pilu/fresh

RUN go mod download

# CMD ["fresh", "-c", ".fresh.conf"]
