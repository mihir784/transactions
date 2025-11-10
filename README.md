# Transactions Service

## How to run

The easiest way is to use Docker. You need Docker Engine & Compose installed.

### Install Docker:
- macOS/Windows: https://www.docker.com/products/docker-desktop/
- Linux: https://docs.docker.com/engine/install

Verify the installation using commands:
```
docker --version
docker compose version
```

### Clone & open the repository
```
git clone https://github.com/mihir784/transactions.git
cd transactions
```

### Build & run the app using docker
Run the command:
```
docker compose up --build
```

Postgres db will run on localhost:5432 & app will listen on 8080 port, this can be configured using docker-compose.yml.
On the first start, app will run the migrations that automatically creates the tables requires (accounts, transactions_log)


## Exposed Endpoint
- '/accounts' - Account Creation Endpoint (POST):
```
curl -i -X POST localhost:8080/accounts -H 'Content-Type: application/json' -d '{"account_id": 123, "initial_balance": "100.23344"}'
```

- '/accounts/{account_id}' - Account Query Endpoint (GET)
```
curl localhost:8080/accounts/123
```

- '/transactions' - Transaction Submission Endpoint (POST)
```
curl -i -X POST localhost:8080/transactions -H 'Content-Type: application/json' 
-d '{"source_account_id": 123, "destination_account_id" : 456, "amount": "100.12345"}'
```


### Local deployment - without docker
- Ensure installation of go & postgres
- Run postgres server on local with user name and password for 'transactions_db'
```
export POSTGRES_URL="postgres://user:password@db:5432/transactions_db"
go mod tidy
go run cmd/transactions/main.go
```