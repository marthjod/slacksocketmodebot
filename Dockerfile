FROM golang as builder

ARG commit

COPY . /build

WORKDIR /build
ENV CGO_ENABLED=0
ENV GOOS=linux

RUN go build \
    -a \
    -ldflags "-X main.commit=${commit} \
              -extldflags \"-static\"" \
    -o /server .


FROM scratch

COPY --from=builder /server /
ENTRYPOINT ["/server"]
