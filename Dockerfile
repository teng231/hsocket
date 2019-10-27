FROM alpine:3.8

RUN mkdir -p /app

WORKDIR /app
COPY hsocket hsocket
COPY wsClient.js wsClient.js
COPY index.html index.html

EXPOSE 8000

CMD ["./hsocket"]