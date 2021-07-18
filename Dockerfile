FROM gibdd_base_image:v1
ARG USERID
ARG GROUPID
USER $USERID:$GROUPID
RUN mkdir -p /home/node/app
WORKDIR /home/node/app
COPY ./tmp/tmp/gibdd /home/node/app

ENTRYPOINT ["./gibdd"]