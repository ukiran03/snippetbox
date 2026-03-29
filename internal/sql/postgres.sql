-- Create the database with modern defaults
CREATE DATABASE snippetbox
    WITH ENCODING = 'UTF8'
    LOCALE = 'en_US.UTF-8';

-- Connect to the new database (psql specific)
\c snippetbox

-- Best Practice: Create a dedicated schema instead of using 'public'
CREATE SCHEMA IF NOT EXISTS snippets;

-- Set the search path (The Postgres equivalent of 'USE')
SET search_path TO snippets, PUBLIC;

-- Make the search_path permanent
ALTER DATABASE snippetbox SET search_path TO snippets, PUBLIC;

-- Verify the search_path
SHOW search_path;

CREATE TABLE snippets (
    id int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title varchar(100) NOT NULL,
    content text NOT NULL,
    created TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires TIMESTAMPTZ NOT NULL,

    -- Constraint to ensure expiration is after creation
    CONSTRAINT check_expiration CHECK (expires > created)
);

-- Index on the created column
CREATE INDEX idx_snippets_created ON snippets (created);

-- Add some dummy records (which we'll use in the next couple of chapters).
INSERT INTO snippets (title, content, created, expires)
VALUES (
    'An old silent pond',
    'An old silent pond...
A frog jumps into the pond,
splash! Silence again.

– Matsuo Bashō',
    now() AT TIME ZONE 'utc',
    (now() AT TIME ZONE 'utc') + INTERVAL '365 days'
);

INSERT INTO snippets (title, content, created, expires)
VALUES (
    'Over the wintry forest',
    'Over the wintry
    forest, winds howl in rage
    with no leaves to blow.

– Natsume Soseki',
    now() AT TIME ZONE 'utc',
    (now() AT TIME ZONE 'utc') + INTERVAL '365 days'
);

INSERT INTO snippets (title, content, created, expires)
VALUES (
    'First autumn morning',
    'First autumn morning
    the mirror I stare into
    shows my father''s face.

– Murakami Kijo',
    now() AT TIME ZONE 'utc',
    (now() AT TIME ZONE 'utc') + INTERVAL '7 days'
);


----- Creatin New user ----
-- 1. Create the role with a secure password
CREATE USER web WITH PASSWORD 'qwe';

-- 2. Revoke all default permissions from the public schema (Security Best Practice)
REVOKE ALL ON SCHEMA public FROM PUBLIC;

-- 3. Grant connection privileges to the database
GRANT CONNECT ON DATABASE snippetbox TO web;

-- 4. Grant usage and specific DML permissions on your schema
GRANT USAGE ON SCHEMA snippets TO web;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA snippets TO web;

-- 5. Crucial: Grant permissions for future tables automatically
ALTER DEFAULT PRIVILEGES IN SCHEMA snippets
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO web;

-- 6. Grant permissions to use sequences (required for identity/auto-increment columns)
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA snippets TO web;

------- Models -------
-- The SQL for your Go Prepare/Exec call
INSERT INTO snippets (title, content, created, expires)
VALUES ($1, $2, now(), now() + ($3 || ' days')::interval)
RETURNING id;

-- Single-record SQL queries
SELECT id, title, content, created, expires FROM snippets
WHERE expires > now() AND id = $1;

-- Multiple-record SQL queries
SELECT id, title, content, created, expires FROM snippets
WHERE expires > now() ORDER BY id DESC LIMIT 10;


-- Setting up the session manager
-- Ensure we are in the correct schema (if not using 'public')
SET search_path TO snippets, public;

CREATE TABLE sessions (
    token char(43) PRIMARY KEY,
    data bytea NOT NULL,
    expiry timestamptz(6) NOT NULL
);

-- Brin index is an alternative for very large session tables,
-- but B-Tree is the standard best practice for typical web apps.
CREATE INDEX sessions_expiry_idx ON sessions (expiry);


---- Authentication
-- Creating a users model
-- 1. Enable the Case-Insensitive Text extension (Best Practice for emails)
CREATE EXTENSION IF NOT EXISTS citext;

-- 2. Create the users table
CREATE TABLE users (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email CITEXT NOT NULL,
    hashed_password CHAR(60) NOT NULL,
    created TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT users_uc_email UNIQUE (email)
);

INSERT INTO users (NAME, email, hashed_password, created)
       VALUES ($1 $2 $3, now())
       RETURNING id

-- password = 'passukiran'
SELECT id, hashed_password FROM users WHERE email = 'ushakiranreddi@gmail.com';

-- User's existence in the DB
SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)

-- Get a user's record given their id
SELECT NAME, email, created FROM users WHERE id = $1
