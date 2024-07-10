# Build stages
FROM golang:alpine as builder

WORKDIR /app

COPY . .

RUN go build -o main main.go

# Run Stages
FROM alpine

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 3000

CMD [ "/app/main" ]
