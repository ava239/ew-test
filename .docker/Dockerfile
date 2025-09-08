FROM golang:1.25-alpine

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.0
RUN apk add --no-cache nano bash postgresql-client shadow build-base

COPY ./ /app
COPY ./.docker/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
WORKDIR /app

RUN go mod tidy
