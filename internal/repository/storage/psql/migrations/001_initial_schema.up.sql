CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS users (
    user_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash CHAR(60) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Создание таблицы пользовательских данных
CREATE TABLE IF NOT EXISTS user_data (
     data_id UUID PRIMARY KEY,
     user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
     data TEXT NOT NULL,
     data_type TEXT NOT NULL,
     created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
     modified_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);