FROM --platform=$BUILDPLATFORM golang:1.21.1-alpine3.18 as builder

RUN apk add --no-cache make ca-certificates gcc musl-dev linux-headers git jq bash

COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum

WORKDIR /app

ADD . .

RUN go mod download

# build lithosphere with the shared go.mod & go.sum files
RUN make lithosphere

FROM alpine:3.18

COPY --from=builder /app/lithosphere /usr/local/bin
COPY --from=builder /app/lithosphere.toml /app/lithosphere.toml
COPY --from=builder /app/migrations /app/migrations

ENV INDEXER_MIGRATIONS_DIR="/app/migrations"

CMD ["lithosphere", "index", "--config", "/app/lithosphere.toml"]
