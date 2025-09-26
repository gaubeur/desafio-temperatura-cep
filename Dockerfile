FROM golang:latest AS build

WORKDIR /app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server

# Expose port 8080 as required by Cloud Run
EXPOSE 8080

FROM scratch

WORKDIR /app

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=build /app/server .

# Define a variável de ambiente para a chave da API de clima
ENV WEATHERAPI_KEY="ebdba81c2d6c44578a534745252509"

# Command to run the executable
ENTRYPOINT ["/app/server"]
