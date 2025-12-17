# Etapa 1: Build (compilação)
FROM golang:1.25.5-alpine AS build

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o api main.go

# Etapa 2: Imagem mínima (runtime)
FROM alpine:latest

RUN apk add --no-cache ca-certificates
RUN apk add --no-cache tzdata

WORKDIR /app

# Copiar o binário da etapa de build
COPY --from=build /app/api /app/api
COPY --from=build /app/.env /app/.env
COPY --from=build /app/docs /app/docs


# Adicionar comando de depuração para verificar se o binário foi copiado corretamente
# RUN ls -l /app


EXPOSE 8080

CMD ["/app/api","nonlocal"]