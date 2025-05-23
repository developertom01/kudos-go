FROM golang:1.24.3-bullseye as builder

WORKDIR /opt/slack

COPY . /opt/slack

RUN go mod download

RUN go build -o slack ./cmd/slack

FROM debian:bullseye-slim AS runner

WORKDIR /opt/slack
COPY --from=builder /opt/slack/slack /opt/slack/slack

CMD [ "./slack" ]



