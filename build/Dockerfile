# syntax=docker/dockerfile:1
FROM golang:1.23 AS builder
WORKDIR /app
ENV GOTOOLCHAIN=auto
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -buildvcs=false -o /bin/timesheet-api ./cmd/api

FROM gcr.io/distroless/base-debian12
WORKDIR /
ENV TZ=Asia/Jakarta
COPY --from=builder /bin/timesheet-api /timesheet-api
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/timesheet-api"]
