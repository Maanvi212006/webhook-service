FROM golang:1.26 AS build
WORKDIR /app
COPY go.mod ./
COPY *.go ./
RUN CGO_ENABLED=0 go build -o webhook-service .

FROM alpine:3.20
COPY --from=build /app/webhook-service /webhook-service
EXPOSE 8080
CMD ["/webhook-service"]
