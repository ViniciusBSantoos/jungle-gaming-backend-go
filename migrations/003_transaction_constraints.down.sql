
ALTER TABLE wager_transactions
DROP CONSTRAINT IF EXISTS opening_internal_only;

ALTER TABLE wager_transactions
DROP CONSTRAINT IF EXISTS external_transaction_required_fields;
