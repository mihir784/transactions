

FROM golang:1.24-alpine AS build
WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/transactions ./cmd/transactions

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /
COPY --from=build /bin/transactions /transactions
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/transactions"]