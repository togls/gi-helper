ARG GO_VERSION=1.18
ARG ALPINE_VERSION=3.15

FROM --platform=${BUILDPLATFORM} golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

RUN apk update && \
    apk add --no-cache \
    ca-certificates \
    git \
    tzdata

WORKDIR /app

ADD . .

RUN go mod download

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 \
    go build --trimpath --ldflags "-w -s" \
    -o /app/gihelper \
    .

FROM alpine:${ALPINE_VERSION}

ENV TZ=Asia/Shanghai

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=builder /app/gihelper /usr/local/bin/gihelper
COPY --from=builder /app/config.json.example /etc/gihelper/config.json

CMD ["/usr/local/bin/gihelper", "-c", "/etc/gihelper/config.json"]