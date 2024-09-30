FROM golang:1.22.5 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM gcr.io/distroless/base

COPY --from=builder /app/main /main

EXPOSE 8080

CMD ["/main"]