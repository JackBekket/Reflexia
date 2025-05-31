FROM golang:latest AS builder
LABEL maintainer="mintyleaf <mintyleafdev@gmail.com>"

WORKDIR /build

COPY go.mod go.sum ./
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -tags=api -o reflexia_api .

FROM alpine
WORKDIR /

COPY --from=builder /build/reflexia_api /reflexia_api
COPY --from=builder /build/project_config /project_config

CMD ["/reflexia_api"]
