FROM golang:1.23

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /ds18b20_exporter

EXPOSE 9101

CMD ["/ds18b20_exporter"]
