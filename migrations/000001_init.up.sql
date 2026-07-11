CREATE TYPE book_condition AS ENUM ('excellent', 'good', 'fair', 'poor', 'damaged');

CREATE TABLE authors
(
    authors_id INT GENERATED ALWAYS AS IDENTITY NOT NULL,
    first_name VARCHAR(100)                     NOT NULL,
    last_name  VARCHAR(100)                     NOT NULL,
    biography  TEXT                             NOT NULL,
    birth_date DATE,

    CONSTRAINT PK_authors_authors_id PRIMARY KEY (authors_id)
);


CREATE TABLE genres
(
    genres_id INT GENERATED ALWAYS AS IDENTITY NOT NULL,
    name      VARCHAR(100)                     NOT NULL,
    parent_id INT,

    CONSTRAINT PK_genres_id PRIMARY KEY (genres_id),

    CONSTRAINT FK_genres_parent
        FOREIGN KEY (parent_id)
            REFERENCES genres (genres_id)
            ON DELETE SET NULL
            ON UPDATE CASCADE
);


CREATE TABLE publishers
(
    publishers_id INT GENERATED ALWAYS AS IDENTITY NOT NULL,
    name          VARCHAR(100)                     NOT NULL,
    address       VARCHAR(100)                     NOT NULL,
    phone         VARCHAR(30)                      NOT NULL,

    CONSTRAINT PK_publishers_publishers_id PRIMARY KEY (publishers_id)
);

CREATE TABLE books
(
    book_id      INT GENERATED ALWAYS AS IDENTITY NOT NULL,
    title        VARCHAR(200)                     NOT NULL,
    isbn         VARCHAR(20)                      NOT NULL,
    year         INT                              NOT NULL,
    publisher_id INT,
    description  TEXT                             NOT NULL,
    cover_image  VARCHAR(255),

    CONSTRAINT PK_book_book_id PRIMARY KEY (book_id),
    CONSTRAINT FK_publisher_id_book
        FOREIGN KEY (publisher_id)
            REFERENCES publishers (publishers_id)
            ON DELETE RESTRICT
            ON UPDATE CASCADE
);


CREATE TABLE book_authors
(
    book_id    INT NOT NULL,
    authors_id INT NOT NULL,

    CONSTRAINT PK_book_authors PRIMARY KEY (book_id, authors_id),

    CONSTRAINT FK_book_authors_book
        FOREIGN KEY (book_id)
            REFERENCES books (book_id)
            ON DELETE CASCADE,

    CONSTRAINT FK_book_authors_author
        FOREIGN KEY (authors_id)
            REFERENCES authors (authors_id)
            ON DELETE CASCADE
);

CREATE TABLE book_genres
(
    book_id  INT NOT NULL,
    genre_id INT NOT NULL,

    CONSTRAINT PK_book_genre PRIMARY KEY (book_id, genre_id),

    CONSTRAINT FK_book_genres_book
        FOREIGN KEY (book_id)
            REFERENCES books (book_id)
            ON DELETE CASCADE,

    CONSTRAINT FK_book_genres_genre
        FOREIGN KEY (genre_id)
            REFERENCES genres (genres_id)
            ON DELETE CASCADE
);

CREATE TABLE book_copies
(
    book_copy_id INT GENERATED ALWAYS AS IDENTITY NOT NULL,
    book_id      INT                              NOT NULL,
    copy_number  INT                              NOT NULL,
    status       VARCHAR(20)                      NOT NULL,
    condition    book_condition                   NOT NULL DEFAULT 'good',
    CONSTRAINT PK_book_copies_book_copy_id PRIMARY KEY (book_copy_id),
    CONSTRAINT FK_book_book_copies_book
        FOREIGN KEY (book_id)
            REFERENCES books (book_id)
            ON DELETE RESTRICT
            ON UPDATE CASCADE
);

CREATE TABLE readers
(
    reader_id     INT GENERATED ALWAYS AS IDENTITY NOT NULL,
    name          VARCHAR(100)                     NOT NULL,
    phone         VARCHAR(30)                      NOT NULL,
    email         VARCHAR(50),
    registered_at TIMESTAMP WITH TIME ZONE         NOT NULL,
    status        VARCHAR(20)                      NOT NULL,
    max_books     INT                              NOT NULL,

    CONSTRAINT PK_readers_reader_id PRIMARY KEY (reader_id),
    CONSTRAINT UQ_readers_phone UNIQUE (phone)
);

CREATE TABLE transactions
(
    transaction_id INT GENERATED ALWAYS AS IDENTITY NOT NULL,
    copy_id        INT                              NOT NULL,
    reader_id      INT                              NOT NULL,
    borrowed_at    TIMESTAMP WITH TIME ZONE         NOT NULL,
    due_date       DATE                             NOT NULL,
    returned_at    TIMESTAMP WITH TIME ZONE,
    status         VARCHAR(20)                      NOT NULL,
    fine           DECIMAL(10, 2),

    CONSTRAINT PK_transactions_transaction_id PRIMARY KEY (transaction_id),
    CONSTRAINT FK_transactions_book_copies_copy_id
        FOREIGN KEY (copy_id)
            REFERENCES book_copies (book_copy_id)
            ON DELETE RESTRICT
            ON UPDATE CASCADE,
    CONSTRAINT FK_transactions_readers_reader_id
        FOREIGN KEY (reader_id)
            REFERENCES readers (reader_id)
            ON DELETE RESTRICT
            ON UPDATE CASCADE
);

CREATE TABLE reservations
(
    reservation_id INT GENERATED ALWAYS AS IDENTITY NOT NULL,
    copy_id        INT                              NOT NULL,
    reader_id      INT                              NOT NULL,
    reserved_at    TIMESTAMP                        NOT NULL,
    expires_at     TIMESTAMP WITH TIME ZONE         NOT NULL,
    status         VARCHAR(20)                      NOT NULL,

    CONSTRAINT PK_reservations_reservation_id PRIMARY KEY (reservation_id),
    CONSTRAINT FK_reservations_book_copies_copy_id
        FOREIGN KEY (copy_id)
            REFERENCES book_copies (book_copy_id)
            ON DELETE RESTRICT
            ON UPDATE CASCADE,
    CONSTRAINT FK_reservations_readers_reader_id
        FOREIGN KEY (reader_id)
            REFERENCES readers (reader_id)
            ON DELETE RESTRICT
            ON UPDATE CASCADE
);

CREATE TABLE reviews
(
    review_id  INT GENERATED ALWAYS AS IDENTITY NOT NULL,
    book_id    INT                              NOT NULL,
    reader_id  INT                              NOT NULL,
    rating     DECIMAL(2, 1)                    NOT NULL CHECK (rating >= 1 AND rating <= 10),
    comment    TEXT,
    created_at TIMESTAMP WITH TIME ZONE         NOT NULL,

    CONSTRAINT PK_reviews_review_id PRIMARY KEY (review_id),
    CONSTRAINT FK_reviews_book_book_id
        FOREIGN KEY (book_id)
            REFERENCES books (book_id)
            ON DELETE RESTRICT
            ON UPDATE CASCADE,
    CONSTRAINT FK_reviews_readers_reader_id
        FOREIGN KEY (reader_id)
            REFERENCES readers (reader_id)
            ON DELETE RESTRICT
            ON UPDATE CASCADE
);

CREATE TABLE users
(
    user_id       INT GENERATED ALWAYS AS IDENTITY NOT NULL,
    email         VARCHAR(100)                     NOT NULL,
    password_hash VARCHAR(100)                     NOT NULL,
    role          VARCHAR(20)                      NOT NULL,
    reader_id     INT,

    CONSTRAINT PK_users_user_id PRIMARY KEY (user_id),
    CONSTRAINT FK_users_readers_reader_id
        FOREIGN KEY (reader_id)
            REFERENCES readers (reader_id)
            ON DELETE RESTRICT
            ON UPDATE CASCADE
);


CREATE TABLE audit_log
(
    audit_log_id  BIGINT GENERATED ALWAYS AS IDENTITY NOT NULL,
    user_id       INT,
    action        VARCHAR(50)                         NOT NULL,
    entity_type   VARCHAR(50)                         NOT NULL, -- 'books', 'authors', 'publishers'
    entity_id     INT                                 NOT NULL,
    log_timestamp TIMESTAMP WITH TIME ZONE            NOT NULL,

    CONSTRAINT PK_audit_log PRIMARY KEY (audit_log_id),
    CONSTRAINT FK_audit_log_user
        FOREIGN KEY (user_id)
            REFERENCES users (user_id) ON DELETE SET NULL
);

CREATE TABLE settings
(
    setting_id    INT GENERATED ALWAYS AS IDENTITY NOT NULL,
    keySettings   VARCHAR(100)                     NOT NULL,
    valueSettings VARCHAR(100)                     NOT NULL,
    description   TEXT                             NOT NULL,

    CONSTRAINT PK_settings_setting_id PRIMARY KEY (setting_id)
);



