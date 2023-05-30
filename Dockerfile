FROM golang:1.20.4-alpine3.18 as builder
WORKDIR /app
COPY . .

RUN go build -o main .

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/main .
COPY app.env .

EXPOSE 8080
CMD ["/app/main"]
