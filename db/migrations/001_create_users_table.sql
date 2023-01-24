-- +migrate Up
CREATE TABLE users (
  id integer PRIMARY KEY,
  user_name text,
  email text,
  created_at datetime,
  updated_at datetime,
  deleted_at datetime
);

-- +migrate Down
DROP TABLE users;
