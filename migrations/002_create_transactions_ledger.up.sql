
CREATE TABLE wager_transactions (
    id UUID PRIMARY KEY,
    wallet_id UUID NOT NULL REFERENCES wallets(id),
    player_id UUID,
    round_id TEXT,
    game_id TEXT,
    reference_external_transaction_id TEXT,
    reference_transaction_id UUID,
    failure_code TEXT,
    result_balance_cents BIGINT,
    provider_id TEXT,
    external_transaction_id TEXT,
    idempotency_key TEXT,
    payload_hash TEXT,
    transaction_type TEXT NOT NULL,
    amount_cents BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL,
    status TEXT NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT transaction_type_valid CHECK (
        transaction_type IN (
            'OPENING', 'BET', 'WIN',
            'LOSS', 'REFUND', 'ROLLBACK'
        )
    ),

    CONSTRAINT transaction_status_valid CHECK (
        status IN (
            'PENDING', 'PENDING_REFERENCE',
            'PROCESSED', 'REJECTED', 'FAILED'
        )
    ),

    CONSTRAINT transaction_amount_valid
        CHECK (amount_cents >= 0),

    CONSTRAINT provider_external_unique
        UNIQUE (provider_id, external_transaction_id)
);

CREATE TABLE wallet_ledger (
    id UUID PRIMARY KEY,
    wallet_id UUID NOT NULL REFERENCES wallets(id),
    transaction_id UUID NOT NULL REFERENCES wager_transactions(id),
    direction TEXT NOT NULL,
    amount_cents BIGINT NOT NULL,
    balance_before_cents BIGINT NOT NULL,
    balance_after_cents BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT ledger_direction_valid
        CHECK (direction IN ('DEBIT', 'CREDIT')),

    CONSTRAINT ledger_amount_positive
        CHECK (amount_cents > 0),

    CONSTRAINT ledger_balance_valid
        CHECK (
            (direction = 'CREDIT'
             AND balance_after_cents = balance_before_cents + amount_cents)
            OR
            (direction = 'DEBIT'
             AND balance_after_cents = balance_before_cents - amount_cents)
        ),

    CONSTRAINT ledger_transaction_unique
        UNIQUE (wallet_id, transaction_id)
);

CREATE UNIQUE INDEX wager_idempotency_unique
ON wager_transactions (idempotency_key)
WHERE idempotency_key IS NOT NULL;

CREATE INDEX wager_wallet_created_idx
ON wager_transactions (wallet_id, created_at);

CREATE INDEX ledger_wallet_created_idx
ON wallet_ledger (wallet_id, created_at);

-- Impede alteracoes e exclusoes no historico financeiro
CREATE OR REPLACE FUNCTION prevent_ledger_changes()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION
        'O historico financeiro nao pode ser alterado ou excluido';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER wallet_ledger_immutable
BEFORE UPDATE OR DELETE ON wallet_ledger
FOR EACH ROW
EXECUTE FUNCTION prevent_ledger_changes();

CREATE TRIGGER wallet_ledger_no_truncate
BEFORE TRUNCATE ON wallet_ledger
FOR EACH STATEMENT
EXECUTE FUNCTION prevent_ledger_changes();

-- Impede valores negativos no saldo anterior e posterior
ALTER TABLE wallet_ledger
ADD CONSTRAINT ledger_balances_non_negative
CHECK (
    balance_before_cents >= 0
    AND balance_after_cents >= 0
);

-- Impede uma transacao de referenciar a si mesma
ALTER TABLE wager_transactions
ADD CONSTRAINT transaction_no_self_reference
CHECK (
    reference_transaction_id IS NULL
    OR reference_transaction_id <> id
);
