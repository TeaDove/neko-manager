# Start by building the application.
FROM golang:1.25-trixie AS build

WORKDIR /src

ENV CGO_ENABLED=1
COPY go.mod go.sum ./
RUN go get ./...

COPY . .

RUN make build

## Now copy it into our base image.
FROM debian:trixie

RUN rm -rf /var/lib/apt/lists/* \
    && apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl ffmpeg
RUN update-ca-certificates
RUN rm -rf /var/lib/apt/lists/*

COPY --from=build /src/bootstrap /

CMD ["/bootstrap"]
