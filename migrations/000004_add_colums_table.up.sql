ALTER TABLE books
ADD COLUMN created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE;

ALTER TABLE transactions
    ADD COLUMN types VARCHAR(20) DEFAULT 'borrow';

ALTER TABLE transactions
    ADD CONSTRAINT chk_transactions_types
        CHECK (types IN ('borrow', 'return', 'overdue', 'lost'));


ALTER TABLE book_copies ADD COLUMN IF NOT EXISTS reader_id INT;
ALTER TABLE book_copies ADD COLUMN IF NOT EXISTS borrowed_at TIMESTAMP WITH TIME ZONE;

ALTER TABLE book_copies ADD CONSTRAINT fk_book_copies_reader
    FOREIGN KEY (reader_id) REFERENCES readers(reader_id) ON DELETE SET NULL;

CREATE INDEX idx_book_copies_reader_id ON book_copies(reader_id);
