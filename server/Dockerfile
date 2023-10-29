FROM golang:1.21-alpine AS build

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod ./
COPY go.sum ./
RUN go mod download
RUN go mod verify

COPY . .
RUN go build -o bin/nola_otel_server ./cmd

RUN adduser -D -g '' -s /bin/false -h /nola_otel_server nola_otel_server

FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /etc/passwd /etc/passwd

COPY --from=build /app/store/postgre/migration /migration
COPY --from=build /app/bin/nola_otel_server /bin/nola_otel_server

USER nola_otel_server

ENTRYPOINT ["/bin/nola_otel_server"]