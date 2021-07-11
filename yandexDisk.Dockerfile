FROM gibdd_base_image:v1
WORKDIR /yadisk
COPY ./yandex-disk-config /home/node/.config/yandex-disk
# RUN curl http://repo.yandex.ru/yandex-disk/yandex-disk_latest_amd64.deb -o ./yandex-disk_latest_amd64.deb &&\
#     dpkg -i yandex-disk_latest_amd64.deb &&\
#     apt update && apt install iputils-ping
RUN chmod -R 777 /yadisk &&\
    chmod -R 777 /home/node/.config/yandex-disk &&\
    apt update && apt install -y wget gnupg2 &&\
    echo "deb http://repo.yandex.ru/yandex-disk/deb/ stable main" | tee -a /etc/apt/sources.list.d/yandex-disk.list > /dev/null &&\
    wget http://repo.yandex.ru/yandex-disk/YANDEX-DISK-KEY.GPG -O- | apt-key add - &&\
    apt update &&\
    apt install -y yandex-disk &&\
    rm -rf /var/lib/apt/lists/*
# RUN apt update && apt install -y iputils-ping
USER 2000
ENTRYPOINT ["yandex-disk", "--no-daemon", "--dir=/yadisk", "--exclude-dirs=notSync,Фотокамера"]
# CMD ["ping", "yandex.ru"]