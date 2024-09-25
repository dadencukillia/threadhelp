FROM golang:1.23.1-alpine3.20
WORKDIR /home/work
COPY go.mod .
COPY go.sum .
RUN apk add libwebp-dev build-base
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o app --ldflags "-s -w"
CMD ./app
