FROM golang:1.27 AS build

LABEL MAINTAINER="espen.wobbes@nhn.no"

ARG VERSION
ARG DATE

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# compile binary
RUN CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION} -X main.buildDate=${DATE}" -o event-bot ./cmd/main.go

# build final image
FROM alpine:3.23

WORKDIR /app

RUN addgroup -g 1000 -S app-group && adduser -u 1000 -S app-user -G app-group

COPY --from=build /app/gslb-event-bot /app/event-bot

# change ownership of directory
RUN chown -R app-user:app-group /app

USER app-user

CMD [ "./event-bot" ]