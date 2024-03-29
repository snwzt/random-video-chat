FROM docker.io/library/golang:1.22 AS build-env
RUN apt update && apt install -y musl-tools
WORKDIR /app
COPY ./ /app
RUN go mod tidy && CC=musl-gcc go build -o bin/rvc cmd/main.go

FROM docker.io/library/alpine:3.19
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /app
COPY --from=build-env /app/bin/rvc /app/bin/rvc
COPY --from=build-env /app/config/ /app/config/
COPY --from=build-env /app/web/ /app/web/
ENTRYPOINT ["/app/bin/rvc"]