CREATE TABLE IF NOT EXISTS redirects (
  code TEXT PRIMARY KEY,
  token TEXT NOT NULL,
  url TEXT NOT NULL,
  active  BOOLEAN NOT NULL CHECK (active IN (0, 1)),
  client_info TEXT NOT NULL,
  created_at TEXT NOT NULL
);