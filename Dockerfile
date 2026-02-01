FROM golang:1.25 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o service ./cmd/app

FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=builder /app/service /service-courier
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/service-courier"]
