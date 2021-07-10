FROM ubuntu:20.04
WORKDIR /app
COPY ./gibdd /app
RUN apt update &&\
    apt install ca-certificates\
    # iputils-ping \
    curl -y &&\
    curl https://curl.se/ca/cacert.pem -o ./CERTIFICATE.pem &&\
    openssl x509 -outform der -in CERTIFICATE.pem -out CERTIFICATE.crt &&\
    cp CERTIFICATE.crt /usr/local/share/ca-certificate &&\
    update-ca-certificates -y &&\
    rm -rf /var/lib/apt/lists/* &&\
    groupadd --gid 2000 node &&\
    useradd --uid 2000 --gid node --shell /bin/bash --create-home node &&\
    mkdir db &&\
    chmod -R 777 /app/
#     useradd john

USER 2000
ENTRYPOINT ["./gibdd"]
# FROM golang:1.15.6 AS build
# WORKDIR /app
# # Копируем все из текущей папки в контейнер
# COPY ./ /app
# # Скачиваем зависимости и билдим
# # RUN CGO_ENABLED=1 GOOS=linux go build -o ./gibdd ./main.go
# RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o ./gibdd ./main.go
# # Запускаем приложение
# CMD ["ping", "yandex.ru"]

# FROM alpine:3.11
# # FROM scratch
# COPY --from=build /app/gibdd /gibdd
# # ENTRYPOINT [ "./gibdd" ]
# # CMD ["./gibdd"]
# CMD [ "ping yandex.ru" ]
# ENTRYPOINT ["/gibdd"]