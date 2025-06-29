FROM golang:1.24-alpine AS build

WORKDIR /app
COPY . .
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64
RUN apk add --no-cache gcc musl-dev
RUN go build -o palpay

FROM golang:1.24-alpine
WORKDIR /app
RUN apk add --no-cache sqlite
COPY --from=build /app/palpay .
COPY ./init_db.sql .
COPY ./login.html .
COPY ./profile.html .
COPY ./set-card.html .
COPY ./styles ./styles
COPY ./scripts ./scripts
COPY ./entrypoint.sh .
USER root
RUN chmod +x /app/entrypoint.sh

EXPOSE 8080
ENTRYPOINT ["/app/entrypoint.sh"]