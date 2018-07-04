FROM golang:1-alpine AS builder

RUN apk add --no-cache git
RUN wget -qO /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64
RUN chmod +x /usr/local/bin/dep

COPY Gopkg.lock Gopkg.toml /go/src/maunium.net/go/mauimageserver/
WORKDIR /go/src/maunium.net/go/mauimageserver
RUN dep ensure -vendor-only

COPY . /go/src/maunium.net/go/mauimageserver
RUN CGO_ENABLED=0 go build -o /usr/bin/mauimageserver


FROM scratch

COPY --from=builder /usr/bin/mauimageserver /usr/bin/mauimageserver

CMD ["/usr/bin/mauimageserver"]
