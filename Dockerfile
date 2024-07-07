FROM golang:1.22.2-alpine AS build

RUN mkdir /app

WORKDIR /app

COPY .  /app

RUN CGO_ENABLED=0 go build -o redisCl . 

RUN chmod +x /app/redisCl

FROM alpine

RUN mkdir /app

WORKDIR /app

COPY --from=build /app/redisCl  /app/

EXPOSE 6666:6666

ENTRYPOINT [ "/app/redisCl" ]