CREATE INDEX idx_books_isbn ON books (isbn);
CREATE INDEX idx_books_publisher ON books (publisher_id);
CREATE INDEX idx_books_year ON books (year);

CREATE INDEX idx_authors_full_name ON authors (last_name, first_name);

CREATE INDEX idx_readers_email ON readers (email);
CREATE INDEX idx_readers_status ON readers (status);

CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_role ON users (role);

CREATE INDEX idx_transactions_reader_status ON transactions (reader_id, status);
CREATE INDEX idx_transactions_copy_status ON transactions (copy_id, status);
CREATE INDEX idx_transactions_borrowed_at ON transactions (borrowed_at);
CREATE INDEX idx_transactions_due_date ON transactions (due_date) WHERE status = 'active';
CREATE INDEX idx_transactions_returned_at ON transactions (returned_at);

CREATE INDEX idx_reservations_reader_status ON reservations (reader_id, status);
CREATE INDEX idx_reservations_copy_status ON reservations (copy_id, status);
CREATE INDEX idx_reservations_expires_at ON reservations (expires_at) WHERE status = 'active';

CREATE INDEX idx_reviews_book_rating ON reviews (book_id, rating);
CREATE INDEX idx_reviews_created_at ON reviews (created_at);

CREATE INDEX idx_book_copies_book_status ON book_copies (book_id, status);

CREATE INDEX idx_audit_entity ON audit_log (entity_type, entity_id);
CREATE INDEX idx_audit_action_timestamp ON audit_log (action, log_timestamp);
CREATE INDEX idx_audit_user_timestamp ON audit_log (user_id, log_timestamp);