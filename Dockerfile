FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod ./
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o bridge .

FROM alpine:3.20
RUN adduser -D -H bridge
COPY --from=build /app/bridge /usr/local/bin/bridge
USER bridge
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/bridge"]
