# Собираем исходник в контейнере
FROM golang:1.16
WORKDIR /go/src/app
COPY . .
RUN go mod download && go build -o /tmp/gibdd ./main.go

CMD ["cp", "-a", "/tmp/gibdd", "/tmp/bin1"]
# RUN cp -a /tmp/healthcheck /tmp/bin1