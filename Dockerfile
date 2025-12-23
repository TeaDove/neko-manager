FROM golang:1.25-trixie AS build

WORKDIR /src

ENV CGO_ENABLED=1
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o=bootstrap main.go

FROM gcr.io/distroless/base-debian13:latest

COPY --from=build /src/bootstrap /

CMD ["/bootstrap"]
