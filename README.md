# Confiapay-BMP

Microserviço que acessa as funcionalidades da API do Banco BMP.



## Descrição

Microserviço de acesso aos serviços bancários do Banco BMP, permite simulação, gravação e consulta de propostas.
Produtos suportados: FGTS, Crédito Pessoal INSS, Crédito CLT.
 


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

- [Documentação da API](https://bmp-fgts.Confiapay.com.br/docs)



