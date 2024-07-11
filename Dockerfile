# Build stages
FROM golang:alpine as builder

WORKDIR /app

COPY . .

RUN go build -o main main.go

RUN apk add curl

RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.1/migrate.linux-amd64.tar.gz | tar xvz

# Run Stages
FROM alpine

WORKDIR /app

COPY --from=builder /app/main .

COPY --from=builder /app/migrate .

COPY .env start.sh wait-for.sh ./

COPY db/migration ./migration

EXPOSE 3000

CMD [ "/app/main" ]

ENTRYPOINT [ "/app/start.sh" ]
