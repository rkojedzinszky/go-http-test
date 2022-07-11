FROM alpine

ARG TARGETARCH

COPY go-http-test.${TARGETARCH} /go-http-test

USER 65534

CMD ["/go-http-test"]
