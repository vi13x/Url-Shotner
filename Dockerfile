FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o url-shortener ./

FROM gcr.io/distroless/static-debian12
WORKDIR /
COPY --from=builder /app/url-shortener /url-shortener
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/url-shortener"]

