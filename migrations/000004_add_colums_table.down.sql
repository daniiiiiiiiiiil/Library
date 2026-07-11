ALTER TABLE transactions
DROP CONSTRAINT IF EXISTS chk_transactions_types;

ALTER TABLE transactions
DROP COLUMN IF EXISTS types;

ALTER TABLE books
DROP COLUMN IF EXISTS updated_at,
DROP COLUMN IF EXISTS created_at,