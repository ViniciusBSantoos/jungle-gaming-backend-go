
ALTER TABLE wager_transactions
ADD CONSTRAINT external_transaction_required_fields
CHECK (
    transaction_type = 'OPENING'
    OR (
        player_id IS NOT NULL
        AND provider_id IS NOT NULL
        AND provider_id <> ''
        AND external_transaction_id IS NOT NULL
        AND external_transaction_id <> ''
        AND idempotency_key IS NOT NULL
        AND idempotency_key <> ''
        AND payload_hash IS NOT NULL
        AND payload_hash <> ''
        AND round_id IS NOT NULL
        AND round_id <> ''
        AND game_id IS NOT NULL
        AND game_id <> ''
    )
);

ALTER TABLE wager_transactions
ADD CONSTRAINT opening_internal_only
CHECK (
    transaction_type <> 'OPENING'
    OR (
        provider_id IS NULL
        AND external_transaction_id IS NULL
        AND idempotency_key IS NULL
        AND payload_hash IS NULL
        AND round_id IS NULL
        AND game_id IS NULL
    )
);
