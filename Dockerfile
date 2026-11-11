FROM golang:1.26-alpine AS build

ARG GOPROXY=https://goproxy.cn,direct
ENV GOPROXY=${GOPROXY}

WORKDIR /src
COPY go.mod ./
COPY controller ./controller
COPY main.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/test-server .

FROM alpine:3.21

RUN adduser -D -u 10001 app
USER app

COPY --from=build /out/test-server /usr/local/bin/test-server

ENV PORT=18080
EXPOSE 18080
ENTRYPOINT ["/usr/local/bin/test-server"]
