CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS pg_trgm;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
        CREATE TYPE user_role AS ENUM ('buyer', 'seller', 'admin');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'listing_status') THEN
        CREATE TYPE listing_status AS ENUM ('active', 'sold', 'archived');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'order_status') THEN
        CREATE TYPE order_status AS ENUM ('pending', 'paid', 'shipped', 'delivered', 'cancelled', 'refunded');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'message_status') THEN
        CREATE TYPE message_status AS ENUM ('sent', 'read');
    END IF;
END $$;

-- Users
CREATE TABLE IF NOT EXISTS users (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email       TEXT UNIQUE NOT NULL,
    password    TEXT NOT NULL,
    name        TEXT NOT NULL,
    role        user_role NOT NULL DEFAULT 'buyer',
    avatar      TEXT,
    bio         TEXT,
    rating      NUMERIC(3,2) DEFAULT 0,
    review_count INT DEFAULT 0,
    stripe_account_id TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Categories
CREATE TABLE IF NOT EXISTS categories (
    id      SERIAL PRIMARY KEY,
    name    TEXT UNIQUE NOT NULL,
    slug    TEXT UNIQUE NOT NULL,
    icon    TEXT
);

-- Listings
CREATE TABLE IF NOT EXISTS listings (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    seller_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id INT REFERENCES categories(id),
    title       TEXT NOT NULL,
    description TEXT NOT NULL,
    price       NUMERIC(12,2) NOT NULL,
    currency    TEXT NOT NULL DEFAULT 'USD',
    status      listing_status NOT NULL DEFAULT 'active',
    location    TEXT,
    images      TEXT[] DEFAULT '{}',
    tags        TEXT[] DEFAULT '{}',
    views       INT DEFAULT 0,
    search_vec  TSVECTOR GENERATED ALWAYS AS (
        to_tsvector('english', title || ' ' || description)
    ) STORED,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS listings_search_idx ON listings USING GIN(search_vec);
CREATE INDEX IF NOT EXISTS listings_seller_idx ON listings(seller_id);
CREATE INDEX IF NOT EXISTS listings_category_idx ON listings(category_id);
CREATE INDEX IF NOT EXISTS listings_status_idx ON listings(status);

-- Orders
CREATE TABLE IF NOT EXISTS orders (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    listing_id          UUID NOT NULL REFERENCES listings(id),
    buyer_id            UUID NOT NULL REFERENCES users(id),
    seller_id           UUID NOT NULL REFERENCES users(id),
    amount              NUMERIC(12,2) NOT NULL,
    platform_fee        NUMERIC(12,2) NOT NULL DEFAULT 0,
    status              order_status NOT NULL DEFAULT 'pending',
    stripe_payment_id   TEXT,
    stripe_transfer_id  TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- Chats
CREATE TABLE IF NOT EXISTS chats (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    listing_id  UUID REFERENCES listings(id),
    buyer_id    UUID NOT NULL REFERENCES users(id),
    seller_id   UUID NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(listing_id, buyer_id, seller_id)
);

-- Messages
CREATE TABLE IF NOT EXISTS messages (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chat_id     UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    sender_id   UUID NOT NULL REFERENCES users(id),
    body        TEXT NOT NULL,
    status      message_status NOT NULL DEFAULT 'sent',
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS messages_chat_idx ON messages(chat_id, created_at);

-- Reviews
CREATE TABLE IF NOT EXISTS reviews (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id    UUID NOT NULL REFERENCES orders(id),
    reviewer_id UUID NOT NULL REFERENCES users(id),
    seller_id   UUID NOT NULL REFERENCES users(id),
    rating      INT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    body        TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(order_id, reviewer_id)
);

-- Refresh tokens
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Notifications
CREATE TABLE IF NOT EXISTS notifications (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type       TEXT NOT NULL,
    payload    JSONB NOT NULL DEFAULT '{}',
    is_read    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS notifications_user_idx ON notifications(user_id, is_read, created_at DESC);

-- Seed categories
INSERT INTO categories (name, slug, icon) VALUES
    ('Electronics',   'electronics',   '💻'),
    ('Clothing',      'clothing',      '👕'),
    ('Home & Garden', 'home-garden',   '🏡'),
    ('Sports',        'sports',        '⚽'),
    ('Vehicles',      'vehicles',      '🚗'),
    ('Books',         'books',         '📚'),
    ('Toys',          'toys',          '🧸'),
    ('Other',         'other',         '📦')
ON CONFLICT DO NOTHING;
