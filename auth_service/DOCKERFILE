FROM golang:1.23.3 AS builder

RUN apt-get update && apt-get install -y git

WORKDIR /app

COPY src/go.mod src/go.sum ./

RUN go mod download

COPY src/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/server .

FROM gcr.io/distroless/static:nonroot

WORKDIR /root/

COPY --from=builder /go/bin/server .

EXPOSE 5678

ENTRYPOINT [ "/root/server" ]
