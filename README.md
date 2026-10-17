
# Jungle Gaming — Backend Go

API REST desenvolvida em Go para processamento seguro de apostas e gerenciamento de carteiras digitais, com autenticação via Keycloak e persistência em PostgreSQL.

## Tecnologias

- Go
- PostgreSQL 16
- Docker Compose
- Keycloak 26.4
- OAuth 2.0 e JWT
- pgx/v5
- Testes automatizados em Go

## Arquitetura

O projeto está organizado em camadas:

- `cmd/api`: inicialização do servidor HTTP.
- `internal/httpapi`: endpoints e tratamento de requisições.
- `internal/auth`: autenticação JWT e autorização por permissões.
- `internal/service`: regras de negócio e processamento de apostas.
- `internal/repository`: acesso ao PostgreSQL.
- `internal/domain`: entidades e regras de domínio.
- `internal/database`: configuração da conexão com o banco.
- `migrations`: estrutura SQL do banco de dados.

## Funcionalidades

- Autenticação com Keycloak e tokens JWT.
- Validação de assinatura, emissor, audiência e expiração.
- Autorização por permissões (`wallet:read` e `wallet:write`).
- Consulta do saldo da carteira do jogador autenticado.
- Processamento transacional de apostas.
- Verificação de saldo suficiente.
- Proteção contra apostas em carteiras de outros jogadores.
- Idempotência para evitar débitos duplicados.
- Controle de concorrência com bloqueios no PostgreSQL.
- Histórico financeiro em ledger.
- Registro de eventos no padrão Transactional Outbox.

## Endpoints

### GET /health

Verifica se o servidor HTTP está respondendo.

Autenticação: não necessária.

Resposta:

```json
{
  "status": "ok"
}
```

### GET /wallet/me

Consulta a carteira BRL vinculada ao usuário autenticado.

Permissão necessária: `wallet:read`.

Exemplo de resposta:

```json
{
  "wallet_id": "11111111-1111-4111-8111-111111111111",
  "player_id": "fe096f62-cfdd-4532-b5ad-8ef8201c5825",
  "currency": "BRL",
  "balance_cents": 8500
}
```

### POST /bets

Registra uma aposta utilizando a carteira do usuário autenticado.

Permissão necessária: `wallet:write`.

Exemplo de requisição:

```json
{
  "provider_id": "provider-01",
  "game_id": "game-01",
  "round_id": "round-01",
  "external_transaction_id": "bet-001",
  "idempotency_key": "request-001",
  "amount_cents": 1500
}
```

Exemplo de resposta:

```json
{
  "transaction_id": "810c8655-d499-4c06-a84b-8c076094bac0",
  "status": "PROCESSED",
  "balance_cents": 8500,
  "replayed": false
}
```

Ao repetir uma requisição com a mesma chave de idempotência e os mesmos dados, a API retorna a transação anterior com `replayed: true`, sem realizar outro débito.

## Configuração do ambiente

Pré-requisitos:

- Go instalado.
- Docker com Docker Compose.
- Git.

1. Copie `.env.example` para `.env`.
2. Configure as senhas locais no `.env`.
3. Inicie PostgreSQL e Keycloak:

```bash
docker compose up -d
```

4. Aplique as migrations SQL na ordem numérica.
5. No Keycloak, configure o realm `jungle-gaming`, o cliente `jungle-backend-api`, as permissões e o usuário de teste.

O Keycloak deve emitir access tokens com a audiência `jungle-backend-api` e as permissões correspondentes.

6. Configure `DATABASE_URL` apontando para o PostgreSQL local, com o banco `jungle_gaming`.

7. Inicie a API:

```bash
go run ./cmd/api
```

Servidor: `http://localhost:8080`

Keycloak: `http://localhost:8081`

Observação: o realm, o cliente e o usuário do Keycloak precisam ser configurados no ambiente local. As configurações do Keycloak não são provisionadas automaticamente pelo Docker Compose atual.

## Testes

Executar todos os testes:

```bash
go test -count=1 ./...
```

Executar com detector de condições de corrida:

```bash
go test -race -count=1 ./...
```

Alguns testes de integração dependem do PostgreSQL local e de dados de teste previamente configurados.

O teste com access token real do Keycloak utiliza a variável `TEST_ACCESS_TOKEN` e é ignorado quando ela não está definida.

## Segurança

- Valores monetários representados em centavos (`int64`).
- Tokens JWT verificados antes do acesso às rotas protegidas.
- Associação da carteira ao usuário autenticado.
- Consultas SQL parametrizadas.
- Transações financeiras com commit e rollback.
- Idempotência e controle de concorrência.
- Segredos locais mantidos fora do Git pelo `.gitignore`.

Este projeto utiliza configurações de desenvolvimento local. Para produção, seriam necessárias configurações adicionais de segurança, gerenciamento de credenciais, observabilidade e publicação dos eventos da Outbox.

## Autor

Vinicius Batista Santos
