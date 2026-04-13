
PRAGMA foreign_keys = ON;

CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    telegram_id TEXT UNIQUE,
    passkey_id TEXT UNIQUE,
    registration_date TEXT DEFAULT (DATETIME('now')),
    deleted_at TEXT
);

CREATE TABLE projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    start_date TEXT,
    end_date TEXT,
    budget INTEGER,
    user_id INTEGER REFERENCES users (id),
    deleted_at TEXT
);

CREATE TABLE cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    credit_limit INTEGER,
    cutoff_date TEXT NOT NULL,
    user_id INTEGER REFERENCES users (id),
    deleted_at TEXT
);

CREATE TABLE monthly_expenses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    expense TEXT NOT NULL,
    amount INTEGER NOT NULL,
    amount_paid INTEGER NOT NULL DEFAULT 0,
    remaining INTEGER GENERATED ALWAYS AS (amount - amount_paid) VIRTUAL,
    paid INTEGER GENERATED ALWAYS AS (CASE WHEN amount_paid >= amount THEN 1 ELSE 0 END) VIRTUAL,
    comment TEXT,
    purchase_msi INTEGER DEFAULT 0,
    credit_card TEXT,
    payment_number INTEGER,
    project_id INTEGER REFERENCES projects (id),
    user_id INTEGER REFERENCES users (id),
    deleted_at TEXT
);

CREATE TABLE monthly_finances (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    month TEXT NOT NULL,
    bank_money INTEGER NOT NULL,
    cash INTEGER NOT NULL,
    expected_income INTEGER NOT NULL
);

CREATE TABLE purchases (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    purchase TEXT NOT NULL,
    date TEXT NOT NULL,
    payment_method TEXT NOT NULL,
    amount INTEGER NOT NULL,
    amount_paid INTEGER NOT NULL DEFAULT 0,
    remaining INTEGER GENERATED ALWAYS AS (amount - amount_paid) VIRTUAL,
    paid INTEGER GENERATED ALWAYS AS (CASE WHEN amount_paid >= amount THEN 1 ELSE 0 END) VIRTUAL,
    carry_over_next_month INTEGER DEFAULT 0,
    comment TEXT,
    card_id INTEGER REFERENCES cards (id),
    project_id INTEGER REFERENCES projects (id),
    user_id INTEGER REFERENCES users (id),
    deleted_at TEXT
);

CREATE TABLE monthly_payments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    purchase_id INTEGER REFERENCES purchases (id),
    month INTEGER NOT NULL,
    amount INTEGER NOT NULL,
    paid INTEGER DEFAULT 0
);

CREATE TABLE advance_payments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    purchase_id INTEGER REFERENCES purchases (id),
    date TEXT NOT NULL,
    amount INTEGER NOT NULL,
    comment TEXT
);

CREATE TABLE change_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    table_name TEXT NOT NULL,
    register_id INTEGER NOT NULL,
    field TEXT NOT NULL,
    action TEXT NOT NULL,
    past_value INTEGER,
    new_value INTEGER,
    change_date TEXT DEFAULT (DATETIME('now')),
    user_id INTEGER REFERENCES users (id),
    motive TEXT
);

CREATE TABLE additional_income (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    description TEXT NOT NULL,
    date TEXT NOT NULL,
    amount INTEGER NOT NULL,
    user_id INTEGER REFERENCES users (id),
    comment TEXT
);