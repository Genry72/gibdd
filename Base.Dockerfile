FROM ubuntu:20.04
WORKDIR /app
# COPY ./gibdd /app
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
    chmod -R 777 /app