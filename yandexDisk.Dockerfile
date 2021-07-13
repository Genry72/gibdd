FROM gibdd_base_image:v1
WORKDIR /yadisk
COPY ./tmp/tmp/yandex-disk-config /home/node/.config/yandex-disk
# RUN curl http://repo.yandex.ru/yandex-disk/yandex-disk_latest_amd64.deb -o ./yandex-disk_latest_amd64.deb &&\
#     dpkg -i yandex-disk_latest_amd64.deb &&\
#     apt update && apt install iputils-ping
RUN apt update && apt install -y wget gnupg2 &&\
    echo "deb http://repo.yandex.ru/yandex-disk/deb/ stable main" | tee -a /etc/apt/sources.list.d/yandex-disk.list > /dev/null &&\
    wget http://repo.yandex.ru/yandex-disk/YANDEX-DISK-KEY.GPG -O- | apt-key add - &&\
    apt update &&\
    apt install -y yandex-disk &&\
    rm -rf /var/lib/apt/lists/* &&\
    chown -R node:node /yadisk &&\
    chown -R node:node /home/node/.config/yandex-disk
# RUN apt update && apt install -y iputils-ping
USER node
ENTRYPOINT ["yandex-disk", "--no-daemon", "--dir=/yadisk", "--exclude-dirs=notSync,Фотокамера"]
# CMD ["ping", "yandex.ru"]