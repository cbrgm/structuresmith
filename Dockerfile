FROM golang:1.21.5-alpine3.17 AS build

WORKDIR /structuresmith

COPY . ./
RUN apk --no-cache add make git curl ca-certificates && make release

FROM alpine:latest

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /structuresmith/bin/structuresmith_linux_amd64 /bin/structuresmith

ENTRYPOINT [ "/bin/structuresmith" ]
