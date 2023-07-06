FROM golang as go

COPY . /app
WORKDIR /app
RUN CGO_ENABLED=0 go build ./cmd/serve

FROM ghcr.io/go-rod/rod:v0.113.4

RUN mkdir /app
WORKDIR /app
COPY --from=go /app/serve ./

EXPOSE 3000

CMD ./serve -p :3000 -t http://localhost:8080