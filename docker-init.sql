-- Create the schema
CREATE SCHEMA IF NOT EXISTS snippets;

-- Set the search path for this session
SET search_path TO snippets, PUBLIC;

-- Make the search_path permanent for the database
ALTER DATABASE snippetbox SET search_path TO snippets, PUBLIC;

-- Create snippets table
CREATE TABLE IF NOT EXISTS snippets (
    id int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title varchar(100) NOT NULL,
    content text NOT NULL,
    created TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires TIMESTAMPTZ NOT NULL,
    CONSTRAINT check_expiration CHECK (expires > created)
);

-- Index on the created column
CREATE INDEX IF NOT EXISTS idx_snippets_created ON snippets (created);

-- Create sessions table
CREATE TABLE IF NOT EXISTS sessions (
    token char(43) PRIMARY KEY,
    data bytea NOT NULL,
    expiry timestamptz(6) NOT NULL
);

-- Index for sessions
CREATE INDEX IF NOT EXISTS sessions_expiry_idx ON sessions (expiry);

-- Enable citext extension for case-insensitive email
CREATE EXTENSION IF NOT EXISTS citext;

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email CITEXT NOT NULL,
    hashed_password CHAR(60) NOT NULL,
    created TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT users_uc_email UNIQUE (email)
);

-- Insert demo user (email: demo@example.com, password: demo)
INSERT INTO users (name, email, hashed_password, created)
VALUES ('Demo User', 'demo@example.com', '$2a$12$o3D9X8DS42YS8kim/IdG3e3zFpdJqHqT.FMi71GfG2xa7qddqvYNq', now())
ON CONFLICT (email) DO NOTHING;

-- Insert demo snippet
INSERT INTO snippets (title, content, created, expires)
SELECT
    'Hello World',
    'fmt.Println("Hello, World!")',
    now() AT TIME ZONE 'utc',
    (now() AT TIME ZONE 'utc') + INTERVAL '365 days'
WHERE NOT EXISTS (
    SELECT 1 FROM snippets WHERE title = 'Hello World'
);
