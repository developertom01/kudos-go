FROM golang:1.24.3-bullseye as builder

WORKDIR /opt/googlechat

COPY . /opt/googlechat

RUN go mod download

RUN go build -o googlechat .

FROM debian:bullseye-slim AS runner

WORKDIR /opt/googlechat
COPY --from=builder /opt/googlechat/googlechat /opt/googlechat/googlechat
COPY --from=builder /opt/googlechat/templates/ /opt/googlechat/templates/

CMD [ "./googlechat" ]