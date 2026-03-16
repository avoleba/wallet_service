# Wallet API

REST API для операций с кошельком: пополнение, списание и получение баланса.

## API

### POST /api/v1/wallet

Выполнить операцию: пополнение (DEPOSIT) или списание (WITHDRAW).

**Тело запроса:**
```json
{
  "walletId": "550e8400-e29b-41d4-a716-446655440000",
  "operationType": "DEPOSIT",
  "amount": 1000
}
```

- `walletId` — UUID кошелька (при первом DEPOSIT кошелёк создаётся автоматически)
- `operationType` — `"DEPOSIT"` или `"WITHDRAW"`
- `amount` — сумма (целое число > 0)

**Ответ 200:**
```json
{
  "walletId": "550e8400-e29b-41d4-a716-446655440000",
  "balance": 1000
}
```

### GET /api/v1/wallets/{WALLET_UUID}

Получить баланс кошелька по UUID.

**Ответ 200:**
```json
{
  "walletId": "550e8400-e29b-41d4-a716-446655440000",
  "balance": 1000
}
```

## Запуск

docker compose up --build
