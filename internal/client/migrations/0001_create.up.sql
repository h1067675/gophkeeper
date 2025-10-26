-- Таблица паролей
CREATE TABLE IF NOT EXISTS gk_passwords (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    server_id INTEGER,
    login TEXT NOT NULL,
    password TEXT NOT NULL,
    domain TEXT NOT NULL,
    created_at DATETIME  NOT NULL,
    changed_at DATETIME,
    deleted BLOB DEFAULT FALSE,
    description TEXT
);

-- Таблица банковских карт
CREATE TABLE IF NOT EXISTS gk_bank_cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    server_id INTEGER,
    bank TEXT,
    number TEXT NOT NULL,
    month INTEGER NOT NULL,
    year INTEGER NOT NULL,
    cvc TEXT NOT NULL,
    holder TEXT,
    created_at DATETIME NOT NULL,
    changed_at DATETIME,
    deleted BLOB DEFAULT FALSE,
    description TEXT
);

-- Таблица текстовых данных
CREATE TABLE IF NOT EXISTS gk_text_data (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    server_id INTEGER,
    name TEXT NOT NULL,
    data TEXT NOT NULL,
    created_at DATETIME  NOT NULL,
    changed_at DATETIME,
    deleted BLOB DEFAULT FALSE,
    description TEXT
);

-- Таблица бинарных данных
CREATE TABLE IF NOT EXISTS gk_binary_data (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    server_id INTEGER,
    name TEXT NOT NULL,
    data TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    changed_at DATETIME,
    deleted BLOB DEFAULT FALSE,
    description TEXT
);

-- Таблица с данными синхронизаций
CREATE TABLE IF NOT EXISTS gk_sync (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sync_time DATETIME NOT NULL
);
