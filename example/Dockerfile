FROM golang:1.18-alpine as build

WORKDIR /app
COPY . .
ENV GO111MODULE=off
RUN go build -o /app

FROM alpine:3.16

WORKDIR /app
COPY --from=build app .
RUN chmod +x app

CMD ["./app"]
EXPOSE 8080/tcp
