FROM golang:1.14 AS builder

RUN wget -O /usr/bin/dep https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64 && \
    chmod +x /usr/bin/dep

WORKDIR /go/src/github.com/nahwinrajan/poc-redis-keys-count

COPY . .
RUN dep ensure -v -vendor-only

RUN make build


FROM alpine:latest AS production
RUN apk update &&\
    apk --no-cache add ca-certificates && \
    rm -rf /var/cache/apk/*

WORKDIR /root/
COPY --from=builder /go/src/github.com/nahwinrajan/poc-redis-keys-count .

EXPOSE 3103
ENTRYPOINT ["./poc-redis-keys-count"]

# # hacking WAY
# FROM alpine:latest
# RUN apk update &&\
#     apk --no-cache add ca-certificates && \
#     rm -rf /var/cache/apk/*

# WORKDIR /root/
# COPY poc-redis-keys-count .

# EXPOSE 3103
# ENTRYPOINT ["./poc-redis-keys-count"]