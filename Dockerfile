FROM golang:alpine

RUN mkdir /build /app
WORKDIR /build
COPY . .
RUN go build -o /app ./cmd/mcauth && rm -rf /build

WORKDIR /app
CMD [ "/app/mcauth" ]

