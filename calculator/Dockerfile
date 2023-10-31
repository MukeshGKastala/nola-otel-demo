FROM golang:1.21.3-alpine AS build

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod ./
COPY go.sum ./
RUN go mod download
RUN go mod verify

COPY . .
RUN go build -o bin/nola_otel_calc .

RUN adduser -D -g '' -s /bin/false -h /nola_otel_calc nola_otel_calc

FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /etc/passwd /etc/passwd

COPY --from=build /app/bin/nola_otel_calc /bin/nola_otel_calc

USER nola_otel_calc

ENTRYPOINT ["/bin/nola_otel_calc"]