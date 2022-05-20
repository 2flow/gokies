FROM golang:1.17 AS builder

# enable support for go modules
ENV GO111MODULE=on

WORKDIR /goApp

COPY ./go.mod ./.
RUN go mod download

ADD ./adapter ./adapter
ADD ./internal ./internal
ADD ./cmd ./cmd

RUN go get
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./build/app ./cmd

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /goApp/app ./
CMD ["./app"]