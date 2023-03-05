FROM golang:1.17.4 AS builder
COPY . stunnel
WORKDIR stunnel

RUN go env -w GOPROXY="https://goproxy.cn,direct"
RUN CGO_ENABLED=0 go build -o /bin/stclient ./cmd/client/main.go
RUN CGO_ENABLED=0 go build -o /bin/stserver ./cmd/server/main.go

FROM scratch as stclient
WORKDIR /0990
WORKDIR bin
COPY --from=builder /bin/stclient .
WORKDIR /0990
CMD ["bin/stclient","-c","config/stclient.json"]

FROM scratch as stserver
WORKDIR /0990
WORKDIR bin
COPY --from=builder /bin/stserver .
WORKDIR /0990
CMD ["bin/stserver","-c","config/stserver.json"]