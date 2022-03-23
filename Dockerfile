FROM golang:1.18-alpine

WORKDIR /bot

COPY go.mod .
COPY go.sum .
COPY main.go .

RUN go mod tidy

RUN go build

CMD [ "./codeexecute", "-token", "OTU1ODM2MTA0NTU5NDYwMzYy.YjndvQ.BJYnvB15GZSelxsW5nVy6O7O0So" ]
