-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS users (
    user_id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash CHAR(60) NOT NULL,
    email VARCHAR(255) UNIQUE,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT (now() AT TIME ZONE 'utc'),
    last_login TIMESTAMP WITHOUT TIME ZONE
);

-- Создание таблицы типов данных
CREATE TABLE IF NOT EXISTS data_types (
    type_id SERIAL PRIMARY KEY,
    type_name VARCHAR(50) UNIQUE NOT NULL
);

-- Заполнение таблицы типов данных начальными значениями
INSERT INTO data_types (type_name) VALUES
    ('login/password'),
    ('text'),
    ('binary'),
    ('card')
ON CONFLICT (type_name) DO NOTHING;

-- Создание основной таблицы данных пользователя
CREATE TABLE IF NOT EXISTS user_data (
    data_id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    type_id INT NOT NULL REFERENCES data_types(type_id) ON DELETE RESTRICT,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT (now() AT TIME ZONE 'utc'),
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT (now() AT TIME ZONE 'utc')
);

-- Создание таблицы для хранения логинов и паролей
CREATE TABLE IF NOT EXISTS login_data (
    data_id INT PRIMARY KEY REFERENCES user_data(data_id) ON DELETE CASCADE,
    login VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    comment TEXT
);

-- Создание таблицы для хранения текстовых данных
CREATE TABLE IF NOT EXISTS text_data (
    data_id INT PRIMARY KEY REFERENCES user_data(data_id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    comment TEXT
);

-- Создание таблицы для хранения бинарных данных
CREATE TABLE IF NOT EXISTS binary_data (
    data_id INT PRIMARY KEY REFERENCES user_data(data_id) ON DELETE CASCADE,
    "binary" BYTEA NOT NULL,
    comment TEXT
);

-- Создание таблицы для хранения данных банковских карт
CREATE TABLE IF NOT EXISTS card_data (
    data_id INT PRIMARY KEY REFERENCES user_data(data_id) ON DELETE CASCADE,
    card_number VARCHAR(19) NOT NULL,
    expiry_date DATE NOT NULL,
    cardholder_name VARCHAR(255) NOT NULL,
    comment TEXT
);
