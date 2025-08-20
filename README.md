# go-dollar-exchange-rate

Projeto em Go composto por dois sistemas:

- **`server.go`**: expõe `GET /cotacao`, busca a cotação do dólar na AwesomeAPI, salva a cotação em SQLite e responde ao cliente em JSON.  
- **`client.go`**: consome o endpoint do servidor com timeout e grava a cotação em `cotacao.txt` no formato `Dólar: {valor}`.

## Requisitos

- **Go 1.20+**

---

## Como Executar

1) **Instalar dependências**
```bash
go mod tidy
```

2) **Subir o servidor** (porta 8080):
```bash
go run server.go
```

3) **Executar o cliente**:
```bash
go run client.go
```

4) **Ver o arquivo gerado**:
```bash
cat cotacao.txt
```

---

## Testes

- Chamada direta ao endpoint:
```bash
curl -s http://localhost:8080/cotacao
```

- Executar o client para gerar/atualizar `cotacao.txt`:
```bash
go run client.go
```
