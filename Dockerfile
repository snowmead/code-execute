FROM golang:1.18-alpine

ARG app
ARG token

WORKDIR /bot

COPY go.mod .
COPY go.sum .
COPY main.go .

RUN go mod tidy

RUN go build

CMD [ "./codeexecute" ]
