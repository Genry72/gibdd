FROM gibdd_base_image:v1
WORKDIR /app
COPY ./gibdd /app
RUN chmod -R 777 /app

USER 2000
ENTRYPOINT ["./gibdd"]