FROM ubuntu:20.04
ARG USERID
ARG GROUPID
RUN groupadd --gid $GROUPID node &&\
    useradd -rm -d /home/node -s /bin/bash -G sudo -u $USERID --gid $GROUPID node &&\
    apt update &&\
    apt install ca-certificates\
    tzdata\
    # sudo \
    # iputils-ping \
    curl -y &&\
    curl https://curl.se/ca/cacert.pem -o ./CERTIFICATE.pem &&\
    openssl x509 -outform der -in CERTIFICATE.pem -out CERTIFICATE.crt &&\
    cp CERTIFICATE.crt /usr/local/share/ca-certificate &&\
    update-ca-certificates -y &&\
    rm -rf /var/lib/apt/lists/* &&\
    ln -fs /usr/share/zoneinfo/Europe/Moscow /etc/localtime && dpkg-reconfigure -f noninteractive tzdata
# RUN echo 'node:node' | chpasswd