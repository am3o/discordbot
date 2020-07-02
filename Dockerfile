FROM golang:1.14.4 AS build

COPY . /app
WORKDIR /app
RUN go build -mod=vendor -a -ldflags '-w' -o main

FROM alpine

COPY resources/dictonary.json /usr/local/bin/resources/dictonary.json
COPY --from=build /app/main /usr/local/bin/app

RUN /usr/local/bin/app
