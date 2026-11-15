FROM golang:1.26-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /app/comments ./cmd/app

FROM alpine:3.22

WORKDIR /app
COPY --from=build /app/comments ./comments
COPY config/config.yml ./config/config.yml

EXPOSE 8080
ENTRYPOINT ["./comments", "-c", "./config/config.yml"]
