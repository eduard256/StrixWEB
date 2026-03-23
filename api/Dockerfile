FROM golang:1.26-alpine AS build

RUN apk add --no-cache gcc musl-dev

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o /app .

FROM alpine:3.22

RUN apk add --no-cache ca-certificates curl

# Download latest cameras.db from StrixCamDB releases
RUN curl -sL https://github.com/eduard256/StrixCamDB/releases/download/latest/cameras.db -o /cameras.db

COPY --from=build /app /app

ENV DB_PATH=/cameras.db
ENV LISTEN=:8080

EXPOSE 8080

CMD ["/app"]
