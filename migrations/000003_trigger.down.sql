DROP TRIGGER IF EXISTS trg_update_reader_books_count_insert ON transactions;
DROP TRIGGER IF EXISTS trg_update_reader_books_count_update ON transactions;
DROP TRIGGER IF EXISTS trg_update_book_copy_status_insert ON transactions;
DROP TRIGGER IF EXISTS trg_update_book_copy_status_update ON transactions;
DROP TRIGGER IF EXISTS trg_calculate_due_date ON transactions;
DROP TRIGGER IF EXISTS trg_update_book_rating_insert ON reviews;
DROP TRIGGER IF EXISTS trg_update_book_rating_update ON reviews;
DROP TRIGGER IF EXISTS trg_update_book_rating_delete ON reviews;
DROP TRIGGER IF EXISTS trg_cancel_expired_reservations ON reservations;
DROP TRIGGER IF EXISTS trg_handle_reservation_copy_status ON reservations;
DROP TRIGGER IF EXISTS trg_release_copy_on_reservation_cancel ON reservations;
DROP TRIGGER IF EXISTS trg_prevent_active_transaction_deletion ON transactions;

DROP FUNCTION IF EXISTS update_reader_books_count() CASCADE;
DROP FUNCTION IF EXISTS update_book_copy_status() CASCADE;
DROP FUNCTION IF EXISTS calculate_due_date() CASCADE;
DROP FUNCTION IF EXISTS update_book_rating() CASCADE;
DROP FUNCTION IF EXISTS cancel_expired_reservations() CASCADE;
DROP FUNCTION IF EXISTS handle_reservation_copy_status() CASCADE;
DROP FUNCTION IF EXISTS release_copy_on_reservation_cancel() CASCADE;
DROP FUNCTION IF EXISTS prevent_active_transaction_deletion() CASCADE;

ALTER TABLE readers DROP COLUMN IF EXISTS books_count;
ALTER TABLE books DROP COLUMN IF EXISTS avg_rating;
ALTER TABLE books DROP COLUMN IF EXISTS reviews_count;