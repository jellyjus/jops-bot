FROM golang:1.25-alpine as builder

WORKDIR /app
COPY . .

RUN go build -o bot ./...

FROM alpine

COPY --from=builder /app/bot /bot

CMD ["/bot"]
EXPOSE 8080