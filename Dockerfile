FROM golang:alpine3.23 AS builder
WORKDIR /app
COPY app/* .

ARG gitHash=undefined
ARG buildVersion=undefined
RUN CGO_ENABLED=0 GOOS=linux go build \
  -o /actions-metrics \
  -ldflags=" \
  -X main.gitHash=$gitHash \
  -X main.buildDate=$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
  -X main.buildVersion=$buildVersion \
  -X main.goVersion=$(go version | awk '{print $3}')" \
  main.go

FROM scratch
COPY --from=builder /actions-metrics /actions-metrics
EXPOSE 3018
ENTRYPOINT ["/actions-metrics"]