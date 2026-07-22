# ---- build stage ----
FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /charli ./cmd/api

# ---- run stage ----
FROM alpine:3.20
RUN adduser -D -u 10001 charli
USER charli
COPY --from=build /charli /usr/local/bin/charli
EXPOSE 8080
ENTRYPOINT ["charli"]
