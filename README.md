# Confiapay-BMP

Microserviço que acessa as funcionalidades da API de Cobranças do Banco BMP.



## Descrição

Microserviço de acesso aos serviços de geração,consulta e cancelamento de cobranças na API do BMP.
 


## Autores

- [@FelipeAugst](https://github.com/FelipeAugst)


## Instalação

Clonando repositório e rodando localmente(considerando um servidor rabbitmq rodando na porta 5672, um banco de dados postgres  e um servidor redis devidamente configurados no arquivo ".env")

### Com criação de imagem Docker


```bash
  git clone https://github.com/gap-sys/bmp
  cd bmp
  make local
```
### Sem criação de imagem Docker

```bash
  git clone https://github.com/gap-sys/bmp
  cd bmp-fgts
  go build main.go
  ./main
```
## Referência

 - [Golang](https://go.dev/)
 - [Fiber](https://docs.gofiber.io/)


## Documentação

- [Documentação da API](https://bmp.confiapay.com.br/docs)



