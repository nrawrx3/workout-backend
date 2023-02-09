-- +migrate Up
ALTER TABLE users
  ADD password_hash text;

-- +migrate Down
ALTER TABLE users
  DROP password_hash;

