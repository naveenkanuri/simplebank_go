FROM golang:1.20.4-alpine3.18 as builder
WORKDIR /app
COPY . .

RUN go build -o main .
RUN apk add curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.0/migrate.linux-amd64.tar.gz | tar xvz

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/migrate ./migrate
COPY app.env .
COPY start.sh .
COPY db/migration ./migration

EXPOSE 8080
CMD ["/app/main"]
ENTRYPOINT ["/app/start.sh"]
