FROM golang:1.23.1-alpine3.20
WORKDIR /home/work
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o app --ldflags "-s -w"
CMD ./app
