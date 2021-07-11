FROM gibdd_base_image:v1
WORKDIR /app
COPY ./gibdd /app

USER 2000
ENTRYPOINT ["./gibdd"]