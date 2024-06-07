FROM golang:1.22.4 AS build
WORKDIR /rrd
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY ./udf /udf
RUN GOOS=linux GOARCH=amd64 go build -o /out/service ./cmd/main/rrd.go

FROM scratch AS bin
COPY --from=build /out/service /
COPY --from=build /udf /udf

ENV SHELL=/bin/sh \
    APP=rrd-service
ENTRYPOINT ["/service"]
EXPOSE 8080