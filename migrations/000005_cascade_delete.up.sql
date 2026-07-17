-- ============================================
-- КАСКАДНОЕ УДАЛЕНИЕ ДЛЯ ВСЕХ СВЯЗАННЫХ ТАБЛИЦ
-- ============================================

-- 1. Удаляем старые ограничения для таблицы books
ALTER TABLE book_copies DROP CONSTRAINT IF EXISTS FK_book_book_copies_book;
ALTER TABLE book_authors DROP CONSTRAINT IF EXISTS FK_book_authors_book;
ALTER TABLE book_genres DROP CONSTRAINT IF EXISTS FK_book_genres_book;
ALTER TABLE reviews DROP CONSTRAINT IF EXISTS FK_reviews_book_book_id;

-- 2. Добавляем новые ограничения с CASCADE
ALTER TABLE book_copies
    ADD CONSTRAINT FK_book_book_copies_book
        FOREIGN KEY (book_id)
            REFERENCES books (book_id)
            ON DELETE CASCADE
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
            ON DELETE CASCADE
            ON UPDATE CASCADE;

-- ============================================
-- КАСКАДНОЕ УДАЛЕНИЕ ДЛЯ ИЗДАТЕЛЕЙ
-- ============================================

-- 3. Удаляем старое ограничение для publishers
ALTER TABLE books DROP CONSTRAINT IF EXISTS FK_publisher_id_book;

-- 4. Добавляем новое ограничение с CASCADE
ALTER TABLE books
    ADD CONSTRAINT FK_publisher_id_book
        FOREIGN KEY (publisher_id)
            REFERENCES publishers (publishers_id)
            ON DELETE CASCADE
            ON UPDATE CASCADE;

-- ============================================
-- КАСКАДНОЕ УДАЛЕНИЕ ДЛЯ АВТОРОВ
-- ============================================

-- 5. Удаляем старое ограничение для authors
ALTER TABLE book_authors DROP CONSTRAINT IF EXISTS FK_book_authors_author;

-- 6. Добавляем новое ограничение с CASCADE
ALTER TABLE book_authors
    ADD CONSTRAINT FK_book_authors_author
        FOREIGN KEY (authors_id)
            REFERENCES authors (authors_id)
            ON DELETE CASCADE
            ON UPDATE CASCADE;

-- ============================================
-- КАСКАДНОЕ УДАЛЕНИЕ ДЛЯ ЖАНРОВ
-- ============================================

-- 7. Удаляем старое ограничение для genres
ALTER TABLE book_genres DROP CONSTRAINT IF EXISTS FK_book_genres_genre;

-- 8. Добавляем новое ограничение с CASCADE
ALTER TABLE book_genres
    ADD CONSTRAINT FK_book_genres_genre
        FOREIGN KEY (genre_id)
            REFERENCES genres (genres_id)
            ON DELETE CASCADE
            ON UPDATE CASCADE;

-- ============================================
-- КАСКАДНОЕ УДАЛЕНИЕ ДЛЯ ЧИТАТЕЛЕЙ
-- ============================================

-- 9. Удаляем старые ограничения для readers
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS FK_transactions_readers_reader_id;
ALTER TABLE reservations DROP CONSTRAINT IF EXISTS FK_reservations_readers_reader_id;
ALTER TABLE reviews DROP CONSTRAINT IF EXISTS FK_reviews_readers_reader_id;
ALTER TABLE users DROP CONSTRAINT IF EXISTS FK_users_readers_reader_id;

-- 10. Добавляем новые ограничения с CASCADE
ALTER TABLE transactions
    ADD CONSTRAINT FK_transactions_readers_reader_id
        FOREIGN KEY (reader_id)
            REFERENCES readers (reader_id)
            ON DELETE CASCADE
            ON UPDATE CASCADE;

ALTER TABLE reservations
    ADD CONSTRAINT FK_reservations_readers_reader_id
        FOREIGN KEY (reader_id)
            REFERENCES readers (reader_id)
            ON DELETE CASCADE
            ON UPDATE CASCADE;

ALTER TABLE reviews
    ADD CONSTRAINT FK_reviews_readers_reader_id
        FOREIGN KEY (reader_id)
            REFERENCES readers (reader_id)
            ON DELETE CASCADE
            ON UPDATE CASCADE;

ALTER TABLE users
    ADD CONSTRAINT FK_users_readers_reader_id
        FOREIGN KEY (reader_id)
            REFERENCES readers (reader_id)
            ON DELETE CASCADE
            ON UPDATE CASCADE;

-- ============================================
-- КАСКАДНОЕ УДАЛЕНИЕ ДЛЯ КОПИЙ КНИГ
-- ============================================

-- 11. Удаляем старые ограничения для book_copies
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS FK_transactions_book_copies_copy_id;
ALTER TABLE reservations DROP CONSTRAINT IF EXISTS FK_reservations_book_copies_copy_id;

-- 12. Добавляем новые ограничения с CASCADE
ALTER TABLE transactions
    ADD CONSTRAINT FK_transactions_book_copies_copy_id
        FOREIGN KEY (copy_id)
            REFERENCES book_copies (book_copy_id)
            ON DELETE CASCADE
            ON UPDATE CASCADE;

ALTER TABLE reservations
    ADD CONSTRAINT FK_reservations_book_copies_copy_id
        FOREIGN KEY (copy_id)
            REFERENCES book_copies (book_copy_id)
            ON DELETE CASCADE
            ON UPDATE CASCADE;

-- ============================================
-- КАСКАДНОЕ УДАЛЕНИЕ ДЛЯ ТРАНЗАКЦИЙ И БРОНЕЙ
-- ============================================

-- 13. Добавляем мягкое удаление для транзакций (опционально)
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;

-- 14. Добавляем мягкое удаление для броней (опционально)
ALTER TABLE reservations ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;

-- ============================================
-- ИНДЕКСЫ ДЛЯ ПОЛЕЙ deleted_at (опционально)
-- ============================================

CREATE INDEX IF NOT EXISTS idx_transactions_deleted_at ON transactions(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_reservations_deleted_at ON reservations(deleted_at) WHERE deleted_at IS NULL;

-- ============================================
-- ВЫВОД ИНФОРМАЦИИ
-- ============================================

DO $$
BEGIN
    RAISE NOTICE 'Миграция 000005_cascade_delete выполнена успешно!';
    RAISE NOTICE 'Теперь удаление книги удаляет все ее копии, отзывы и связи';
    RAISE NOTICE 'Удаление издателя удаляет все его книги';
    RAISE NOTICE 'Удаление читателя удаляет все его транзакции и брони';
END $$;