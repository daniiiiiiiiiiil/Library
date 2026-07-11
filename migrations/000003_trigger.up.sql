-- 1. ОБНОВЛЕНИЕ КОЛИЧЕСТВА КНИГ У ЧИТАТЕЛЯ

ALTER TABLE readers ADD COLUMN IF NOT EXISTS books_count INT DEFAULT 0 CHECK (books_count >= 0);

CREATE OR REPLACE FUNCTION update_reader_books_count()
RETURNS TRIGGER AS $$
DECLARE
v_reader_id INT;
    v_change INT;
BEGIN
    IF TG_OP = 'INSERT' THEN
        v_reader_id := NEW.reader_id;

        IF NEW.status = 'borrowed' THEN
            v_change := 1;
        ELSIF NEW.status = 'returned' THEN
            v_change := -1;
ELSE
            RETURN NEW;
END IF;

    ELSIF TG_OP = 'UPDATE' THEN
        IF OLD.status = 'borrowed' AND NEW.status = 'returned' THEN
            v_reader_id := NEW.reader_id;
            v_change := -1;
        ELSIF OLD.status = 'returned' AND NEW.status = 'borrowed' THEN
            v_reader_id := NEW.reader_id;
            v_change := 1;
ELSE
            RETURN NEW;
END IF;
END IF;

    IF (SELECT books_count FROM readers WHERE reader_id = v_reader_id) + v_change < 0 THEN
        RAISE EXCEPTION 'У читателя % не может быть отрицательное количество книг', v_reader_id;
END IF;

UPDATE readers
SET books_count = books_count + v_change
WHERE reader_id = v_reader_id;

RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_reader_books_count_insert
    AFTER INSERT ON transactions
    FOR EACH ROW
    EXECUTE FUNCTION update_reader_books_count();

CREATE TRIGGER trg_update_reader_books_count_update
    AFTER UPDATE ON transactions
    FOR EACH ROW
    EXECUTE FUNCTION update_reader_books_count();


-- 2. ОБНОВЛЕНИЕ СТАТУСА КНИГИ

CREATE OR REPLACE FUNCTION update_book_copy_status()
RETURNS TRIGGER AS $$
DECLARE
v_current_status VARCHAR(20);
    v_new_status VARCHAR(20);
BEGIN
SELECT status INTO v_current_status
FROM book_copies
WHERE book_copy_id = NEW.copy_id
    FOR UPDATE;

IF NOT FOUND THEN
        RAISE EXCEPTION 'Копия книги с ID % не найдена', NEW.copy_id;
END IF;

    IF TG_OP = 'INSERT' THEN
        IF NEW.status = 'borrowed' THEN
            IF v_current_status != 'available' THEN
                RAISE EXCEPTION 'Книга #% не доступна для выдачи (статус: %)',
                                NEW.copy_id, v_current_status;
END IF;
            v_new_status := 'borrowed';

        ELSIF NEW.status = 'returned' THEN
            IF v_current_status = 'borrowed' THEN
                v_new_status := 'available';
ELSE
                RAISE EXCEPTION 'Книга #% не была выдана (статус: %)',
                                NEW.copy_id, v_current_status;
END IF;
END IF;

    ELSIF TG_OP = 'UPDATE' THEN
        IF OLD.status = 'borrowed' AND NEW.status = 'returned' THEN
            v_new_status := 'available';
        ELSIF OLD.status = 'returned' AND NEW.status = 'borrowed' THEN
            IF v_current_status = 'available' THEN
                v_new_status := 'borrowed';
ELSE
                RAISE EXCEPTION 'Книга #% не доступна для выдачи', NEW.copy_id;
END IF;
END IF;
END IF;

    IF v_new_status IS NOT NULL THEN
UPDATE book_copies
SET status = v_new_status
WHERE book_copy_id = NEW.copy_id;
END IF;

RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_book_copy_status_insert
    AFTER INSERT ON transactions
    FOR EACH ROW
    EXECUTE FUNCTION update_book_copy_status();

CREATE TRIGGER trg_update_book_copy_status_update
    AFTER UPDATE ON transactions
    FOR EACH ROW
    EXECUTE FUNCTION update_book_copy_status();


-- 3. РАСЧЕТ ДАТЫ ВОЗВРАТА

CREATE OR REPLACE FUNCTION calculate_due_date()
RETURNS TRIGGER AS $$
DECLARE
v_loan_days INT := 14;
BEGIN
    IF NEW.status != 'borrowed' THEN
        RETURN NEW;
END IF;

    NEW.borrowed_at := CURRENT_TIMESTAMP;
    NEW.due_date := CURRENT_DATE + v_loan_days;

RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_calculate_due_date
    BEFORE INSERT ON transactions
    FOR EACH ROW
    EXECUTE FUNCTION calculate_due_date();


-- 4. ОБНОВЛЕНИЕ РЕЙТИНГА КНИГИ

ALTER TABLE books ADD COLUMN IF NOT EXISTS avg_rating DECIMAL(3, 2);
ALTER TABLE books ADD COLUMN IF NOT EXISTS reviews_count INT DEFAULT 0;

CREATE OR REPLACE FUNCTION update_book_rating()
RETURNS TRIGGER AS $$
DECLARE
v_book_id INT;
    v_avg_rating DECIMAL(3, 2);
    v_reviews_count INT;
BEGIN
    IF TG_OP = 'DELETE' THEN
        v_book_id := OLD.book_id;
ELSE
        v_book_id := NEW.book_id;
END IF;

SELECT
    ROUND(AVG(rating)::NUMERIC, 2),
    COUNT(*)
INTO
    v_avg_rating,
    v_reviews_count
FROM reviews
WHERE book_id = v_book_id;

-- Обновляем книгу
UPDATE books
SET
    avg_rating = v_avg_rating,
    reviews_count = v_reviews_count
WHERE book_id = v_book_id;

RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_book_rating_insert
    AFTER INSERT ON reviews
    FOR EACH ROW
    EXECUTE FUNCTION update_book_rating();

CREATE TRIGGER trg_update_book_rating_update
    AFTER UPDATE ON reviews
    FOR EACH ROW
    EXECUTE FUNCTION update_book_rating();

CREATE TRIGGER trg_update_book_rating_delete
    AFTER DELETE ON reviews
    FOR EACH ROW
    EXECUTE FUNCTION update_book_rating();


-- 5. СТАТУС РЕЗЕРВАЦИЙ (АВТОМАТИЧЕСКАЯ ОТМЕНА ПРОСРОЧЕННЫХ)

CREATE OR REPLACE FUNCTION cancel_expired_reservations()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'active' AND NEW.expires_at < CURRENT_TIMESTAMP THEN
UPDATE reservations
SET status = 'cancelled'
WHERE reservation_id = NEW.reservation_id;

UPDATE book_copies
SET status = 'available'
WHERE book_copy_id = NEW.copy_id;
END IF;

RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_cancel_expired_reservations
    AFTER INSERT OR UPDATE ON reservations
                        FOR EACH ROW
                        EXECUTE FUNCTION cancel_expired_reservations();


-- 6. ОБНОВЛЕНИЕ СТАТУСА КОПИИ ПРИ РЕЗЕРВАЦИИ

CREATE OR REPLACE FUNCTION handle_reservation_copy_status()
RETURNS TRIGGER AS $$
DECLARE
v_current_status VARCHAR(20);
BEGIN
    IF NEW.status != 'active' THEN
        RETURN NEW;
END IF;

SELECT status INTO v_current_status
FROM book_copies
WHERE book_copy_id = NEW.copy_id
    FOR UPDATE;

IF NOT FOUND THEN
        RAISE EXCEPTION 'Копия книги с ID % не найдена', NEW.copy_id;
END IF;

    IF v_current_status NOT IN ('available', 'reserved') THEN
        RAISE EXCEPTION 'Книга #% не доступна для резервации (статус: %)',
                        NEW.copy_id, v_current_status;
END IF;

    IF v_current_status = 'reserved' THEN
        PERFORM 1 FROM reservations
        WHERE copy_id = NEW.copy_id
          AND status = 'active'
          AND reader_id != NEW.reader_id;

        IF FOUND THEN
            RAISE EXCEPTION 'Книга #% уже зарезервирована другим читателем', NEW.copy_id;
END IF;
END IF;

UPDATE book_copies
SET status = 'reserved'
WHERE book_copy_id = NEW.copy_id;

RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_handle_reservation_copy_status
    BEFORE INSERT ON reservations
    FOR EACH ROW
    EXECUTE FUNCTION handle_reservation_copy_status();



-- 7. ОСВОБОЖДЕНИЕ КНИГИ ПРИ ОТМЕНЕ РЕЗЕРВАЦИИ

CREATE OR REPLACE FUNCTION release_copy_on_reservation_cancel()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.status = 'active' AND NEW.status = 'cancelled' THEN
UPDATE book_copies
SET status = 'available'
WHERE book_copy_id = NEW.copy_id;
END IF;

RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_release_copy_on_reservation_cancel
    AFTER UPDATE ON reservations
    FOR EACH ROW
    EXECUTE FUNCTION release_copy_on_reservation_cancel();


-- 8. ЗАЩИТА ОТ УДАЛЕНИЯ АКТИВНЫХ ТРАНЗАКЦИЙ

CREATE OR REPLACE FUNCTION prevent_active_transaction_deletion()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.status IN ('borrowed', 'active') THEN
        RAISE EXCEPTION 'Невозможно удалить активную транзакцию #%', OLD.transaction_id;
END IF;
RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_prevent_active_transaction_deletion
    BEFORE DELETE ON transactions
    FOR EACH ROW
    EXECUTE FUNCTION prevent_active_transaction_deletion();