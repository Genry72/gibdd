FROM ubuntu:20.04
WORKDIR /app
COPY ./gibdd /app
# RUN apt update &&\
#     apt install curl -y &&\
#     curl https://curl.se/ca/cacert.pem -o /etc/pki/trust/anchors/cacert.pem &&\
#     sudo update-ca-certificates &&\
#     rm -rf /var/lib/apt/lists/*
RUN update-ca-certificates
ENTRYPOINT ["./gibdd"]
# FROM golang:1.15.6 AS build
# WORKDIR /app
# # Копируем все из текущей папки в контейнер
# COPY ./ /app
# # Скачиваем зависимости и билдим
# # RUN CGO_ENABLED=1 GOOS=linux go build -o ./gibdd ./main.go
# RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o ./gibdd ./main.go
# # Запускаем приложение
# # ENTRYPOINT ping yandex.ru

# FROM alpine:3.11
# # FROM scratch
# COPY --from=build /app/gibdd /gibdd
# # ENTRYPOINT [ "./gibdd" ]
# # CMD ["./gibdd"]
# # CMD [ "ping yandex.ru" ]
# ENTRYPOINT ["/gibdd"]