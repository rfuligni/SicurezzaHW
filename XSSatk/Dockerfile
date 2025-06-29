FROM golang:1.24

WORKDIR /app
COPY . .
RUN go build -o attacker-server main.go

EXPOSE 8090
CMD ["./attacker-server"]