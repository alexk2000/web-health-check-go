FROM golang:1.16-alpine as builder

RUN mkdir -p /app/

WORKDIR /app

COPY cmd ./cmd
COPY pkg ./pkg
COPY go.mod .
COPY go.sum .

RUN go mod download

RUN go build cmd/web-health-check.go

FROM alpine:3.14

LABEL maintainer="alexk2000"

RUN addgroup -g 1000 -S app \
    && adduser -u 1000 -S -G app app

RUN apk add --no-cache tzdata

WORKDIR /home/app

COPY --from=builder /app/web-health-check .
RUN chown -R app:app ./

USER app

ARG VERSION
ENV VERSION=$VERSION

CMD ["./web-health-check"]
