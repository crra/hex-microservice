CREATE TABLE IF NOT EXISTS redirects (
  code TEXT PRIMARY KEY,
  token TEXT NOT NULL,
  url TEXT NOT NULL,
  client_info TEXT NOT NULL,
  created_at TEXT NOT NULL
);