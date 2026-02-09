FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o registry ./cmd/registry

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /app/registry .
EXPOSE 8080
CMD ["/app/registry"]
