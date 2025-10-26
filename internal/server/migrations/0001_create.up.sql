-- 0001_create.up.sql (исправленный рекомендуемый вариант)

CREATE TABLE IF NOT EXISTS gk_users (
    id SERIAL PRIMARY KEY,
    login VARCHAR(254) NOT NULL,
    email VARCHAR(254) NOT NULL,
    password VARCHAR(254) NOT NULL,
    secret_password_hash VARCHAR(254), 
    saltb64 VARCHAR(254) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    confirmed BOOLEAN,
    deleted BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS gk_sessions (
    id SERIAL PRIMARY KEY,
    session VARCHAR(254) UNIQUE,
    allowed_totp BOOLEAN,
    key_totp VARCHAR(254),
    code_email VARCHAR(254),
    expire_code_email TIMESTAMPTZ,
    expire TIMESTAMPTZ NOT NULL,
    users_id INT NOT NULL
);

CREATE TABLE IF NOT EXISTS gk_passwords (
    id SERIAL PRIMARY KEY,
    users_id INT NOT NULL,
    login TEXT NOT NULL,
    password TEXT NOT NULL,
    domain TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    changed_at TIMESTAMPTZ,
    deleted BOOLEAN,
    description TEXT
);

CREATE TABLE IF NOT EXISTS gk_bank_cards (
    id SERIAL PRIMARY KEY,
    users_id INT NOT NULL,
    bank TEXT,
    number TEXT NOT NULL,
    month INT NOT NULL,
    year INT NOT NULL,
    cvc TEXT NOT NULL,
    holder TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    changed_at TIMESTAMPTZ,
    deleted BOOLEAN,
    description TEXT
);

CREATE TABLE IF NOT EXISTS gk_text_data (
    id SERIAL PRIMARY KEY,
    users_id INT NOT NULL,
    name TEXT NOT NULL,
    data TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    changed_at TIMESTAMPTZ,
    deleted BOOLEAN,
    description TEXT
);

CREATE TABLE IF NOT EXISTS gk_binary_data (
    id SERIAL PRIMARY KEY,
    users_id INT NOT NULL,
    name TEXT NOT NULL,
    data TEXT NOT NULL,  
    created_at TIMESTAMPTZ DEFAULT now(),
    changed_at TIMESTAMPTZ,
    deleted BOOLEAN,
    description TEXT
);
