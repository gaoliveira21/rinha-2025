FROM golang:1.24-alpine AS build

WORKDIR /usr/app

COPY . /usr/app

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server.go

FROM scratch

WORKDIR /usr/app

COPY --from=build /usr/app/server /usr/app/server

EXPOSE ${PORT}

CMD ["/usr/app/server"]