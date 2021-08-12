FROM golang:1.16 AS build

COPY . /app
WORKDIR /app
RUN go mod vendor
RUN CGO_ENABLED=0 go build -mod=vendor -a -ldflags '-w' -o /app/main

FROM scratch

COPY --from=build /app/main /usr/local/bin/app
CMD /usr/local/bin/app
