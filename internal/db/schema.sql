CREATE TABLE IF NOT EXISTS users (
    id               BIGSERIAL PRIMARY KEY,
    email            TEXT UNIQUE NOT NULL,
    password_hash    TEXT NOT NULL,
    role             TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    confirmed        BOOLEAN NOT NULL DEFAULT false,
    confirm_code     TEXT,
    reset_code       TEXT,
    reset_expires_at TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         BIGSERIAL PRIMARY KEY,
    token_hash TEXT UNIQUE NOT NULL,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS products (
    id             BIGSERIAL PRIMARY KEY,
    ref            TEXT UNIQUE NOT NULL,
    name           TEXT NOT NULL,
    description    TEXT NOT NULL DEFAULT '',
    category       TEXT NOT NULL,
    price_ht_cents BIGINT NOT NULL CHECK (price_ht_cents >= 0),
    vat_percent    INTEGER NOT NULL DEFAULT 20 CHECK (vat_percent >= 0),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS carts (
    id         BIGSERIAL PRIMARY KEY,
    ref        TEXT UNIQUE NOT NULL,
    user_id    BIGINT UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS cart_items (
    cart_id    BIGINT NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity   INTEGER NOT NULL CHECK (quantity > 0),
    PRIMARY KEY (cart_id, product_id)
);

CREATE TABLE IF NOT EXISTS orders (
    id            BIGSERIAL PRIMARY KEY,
    ref           TEXT UNIQUE NOT NULL,
    user_id       BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status        TEXT NOT NULL DEFAULT 'pending'
    CHECK (status IN ('pending', 'paid', 'shipping', 'delivered', 'cancelled')),
    cancel_reason TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS order_items (
    order_id             BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id           BIGINT NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    quantity             INTEGER NOT NULL CHECK (quantity > 0),
    unit_price_ttc_cents BIGINT NOT NULL CHECK (unit_price_ttc_cents >= 0),
    PRIMARY KEY          (order_id, product_id)
);

CREATE TABLE IF NOT EXISTS payments (
    id           BIGSERIAL PRIMARY KEY,
    order_id     BIGINT UNIQUE NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    card_last4   TEXT NOT NULL CHECK (card_last4 ~ '^[0-9]{4}$'),
    amount_cents BIGINT NOT NULL CHECK (amount_cents >= 0),
    paid_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);