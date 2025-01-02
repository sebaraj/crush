FROM golang:1.23.3 AS builder

RUN apt-get update && apt-get install -y git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY main.go ./

COPY ./server ./server

RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/server .

FROM gcr.io/distroless/static:nonroot

WORKDIR /root/

COPY --from=builder /go/bin/server .

EXPOSE 6000

ENTRYPOINT [ "/root/server" ]

