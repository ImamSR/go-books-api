CREATE TABLE IF NOT EXISTS books (
  id          TEXT PRIMARY KEY,
  name        TEXT NOT NULL,
  author      TEXT,
  publisher   TEXT,
  page_count  INT  NOT NULL DEFAULT 0 CHECK (page_count >= 0),
  read_page   INT  NOT NULL DEFAULT 0 CHECK (read_page >= 0 AND read_page <= page_count),
  reading     BOOLEAN NOT NULL DEFAULT FALSE,
  finished    BOOLEAN NOT NULL DEFAULT FALSE,
  inserted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- index utk filter cepat
CREATE INDEX IF NOT EXISTS idx_books_reading  ON books (reading);
CREATE INDEX IF NOT EXISTS idx_books_finished ON books (finished);
-- index utk pencarian nama non-case-sensitive
CREATE INDEX IF NOT EXISTS idx_books_name_lower ON books ((lower(name)));
