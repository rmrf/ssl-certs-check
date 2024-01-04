FROM golang:1.21-alpine as deps
ARG ENV_GOPROXY 
ENV GOPROXY ${ENV_GOPROXY}
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download -x

FROM golang:1.21-alpine as builder
RUN apk --no-cache add git
WORKDIR /app
COPY --from=deps /go/pkg /go/pkg/
COPY . .
ENV CGO_ENABLED=0
ENV GO111MODULE=on
RUN go build -o /app/ssl-certs-check \
    -ldflags "-w -extldflags \"-static\" -X \"main.version=$(git rev-parse HEAD)\"" \
    /app

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app ./
ENTRYPOINT ["./ssl-certs-check"]
EXPOSE 8080