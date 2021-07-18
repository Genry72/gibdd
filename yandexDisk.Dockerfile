FROM gibdd_base_image:v1
ARG USERID
ARG GROUPID
COPY ./tmp/tmp/yandex-disk-config /home/node/.config/yandex-disk

RUN apt update && apt install -y wget gnupg2 &&\
    echo "deb http://repo.yandex.ru/yandex-disk/deb/ stable main" | tee -a /etc/apt/sources.list.d/yandex-disk.list > /dev/null &&\
    wget http://repo.yandex.ru/yandex-disk/YANDEX-DISK-KEY.GPG -O- | apt-key add - &&\
    apt update &&\
    apt install -y yandex-disk &&\
    rm -rf /var/lib/apt/lists/*
# RUN apt update && apt install -y iputils-ping
USER $USERID:$GROUPID
WORKDIR /home/node/yadisk
ENTRYPOINT ["yandex-disk", "--no-daemon", "--dir=/home/node/yadisk", "--exclude-dirs=notSync,Фотокамера"]
# CMD ["ping", "yandex.ru"]