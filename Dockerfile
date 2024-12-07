FROM golang:latest

WORKDIR /app
COPY ./src ./src

WORKDIR /app/src
RUN chmod +x ./build.sh
RUN sh ./build.sh

CMD cp -a ./**/ /app/out/
