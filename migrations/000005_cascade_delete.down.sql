-- ============================================
-- ОТКАТ КАСКАДНОГО УДАЛЕНИЯ
-- ============================================

-- Удаляем новые ограничения
ALTER TABLE book_copies DROP CONSTRAINT IF EXISTS FK_book_book_copies_book;
ALTER TABLE book_authors DROP CONSTRAINT IF EXISTS FK_book_authors_book;
ALTER TABLE book_genres DROP CONSTRAINT IF EXISTS FK_book_genres_book;
ALTER TABLE reviews DROP CONSTRAINT IF EXISTS FK_reviews_book_book_id;
ALTER TABLE books DROP CONSTRAINT IF EXISTS FK_publisher_id_book;
ALTER TABLE book_authors DROP CONSTRAINT IF EXISTS FK_book_authors_author;
ALTER TABLE book_genres DROP CONSTRAINT IF EXISTS FK_book_genres_genre;
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS FK_transactions_readers_reader_id;
ALTER TABLE reservations DROP CONSTRAINT IF EXISTS FK_reservations_readers_reader_id;
ALTER TABLE reviews DROP CONSTRAINT IF EXISTS FK_reviews_readers_reader_id;
ALTER TABLE users DROP CONSTRAINT IF EXISTS FK_users_readers_reader_id;
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS FK_transactions_book_copies_copy_id;
ALTER TABLE reservations DROP CONSTRAINT IF EXISTS FK_reservations_book_copies_copy_id;

-- Возвращаем старые ограничения с RESTRICT
ALTER TABLE book_copies
    ADD CONSTRAINT FK_book_book_copies_book
        FOREIGN KEY (book_id)
            REFERENCES books (book_id)
            ON DELETE RESTRICT
            ON UPDATE CASCADE;

ALTER TABLE book_authors
    ADD CONSTRAINT FK_book_authors_book
        FOREIGN KEY (book_id)
            REFERENCES books (book_id)
            ON DELETE CASCADE
            ON UPDATE CASCADE;

ALTER TABLE book_genres
    ADD CONSTRAINT FK_book_genres_book
        FOREIGN KEY (book_id)
            REFERENCES books (book_id)
            ON DELETE CASCADE
            ON UPDATE CASCADE;

ALTER TABLE reviews
    ADD CONSTRAINT FK_reviews_book_book_id
        FOREIGN KEY (book_id)
            REFERENCES books (book_id)
            ON DELETE RESTRICT
            ON UPDATE CASCADE;

ALTER TABLE books
    ADD CONSTRAINT FK_publisher_id_book
        FOREIGN KEY (publisher_id)
            REFERENCES publishers (publishers_id)
            ON DELETE RESTRICT
            ON UPDATE CASCADE;

ALTER TABLE book_authors
    ADD CONSTRAINT FK_book_authors_author
        FOREIGN KEY (authors_id)
            REFERENCES authors (authors_id)
            ON DELETE CASCADE
            ON UPDATE CASCADE;

ALTER TABLE book_genres
    ADD CONSTRAINT FK_book_genres_genre
        FOREIGN KEY (genre_id)
            REFERENCES genres (genres_id)
            ON DELETE CASCADE
            ON UPDATE CASCADE;

ALTER TABLE transactions
    ADD CONSTRAINT FK_transactions_readers_reader_id
        FOREIGN KEY (reader_id)
            REFERENCES readers (reader_id)
            ON DELETE RESTRICT
            ON UPDATE CASCADE;

ALTER TABLE reservations
    ADD CONSTRAINT FK_reservations_readers_reader_id
        FOREIGN KEY (reader_id)
            REFERENCES readers (reader_id)
            ON DELETE RESTRICT
            ON UPDATE CASCADE;

ALTER TABLE reviews
    ADD CONSTRAINT FK_reviews_readers_reader_id
        FOREIGN KEY (reader_id)
            REFERENCES readers (reader_id)
            ON DELETE RESTRICT
            ON UPDATE CASCADE;

ALTER TABLE users
    ADD CONSTRAINT FK_users_readers_reader_id
        FOREIGN KEY (reader_id)
            REFERENCES readers (reader_id)
            ON DELETE RESTRICT
            ON UPDATE CASCADE;

ALTER TABLE transactions
    ADD CONSTRAINT FK_transactions_book_copies_copy_id
        FOREIGN KEY (copy_id)
            REFERENCES book_copies (book_copy_id)
            ON DELETE RESTRICT
            ON UPDATE CASCADE;

ALTER TABLE reservations
    ADD CONSTRAINT FK_reservations_book_copies_copy_id
        FOREIGN KEY (copy_id)
            REFERENCES book_copies (book_copy_id)
            ON DELETE RESTRICT
            ON UPDATE CASCADE;

-- Удаляем поля deleted_at
ALTER TABLE transactions DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE reservations DROP COLUMN IF EXISTS deleted_at;

-- Удаляем индексы
DROP INDEX IF EXISTS idx_transactions_deleted_at;
DROP INDEX IF EXISTS idx_reservations_deleted_at;

DO $$
BEGIN
    RAISE NOTICE 'Откат миграции 000005_cascade_delete выполнен успешно!';
    RAISE NOTICE 'Восстановлены старые ограничения с RESTRICT';
END $$;