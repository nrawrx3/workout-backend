-- +migrate Up
CREATE TABLE user_sessions (
  id integer PRIMARY KEY,
  created_at datetime,
  updated_at datetime,
  deleted_at datetime,
  user_id integer,
  user_agent text,
  expires_at datetime,
  raw_session_data blob,
  FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- +migrate Down
DROP TABLE user_sessions;

