-- +migrate Up
CREATE UNIQUE INDEX unique_users__email ON users (email);

-- +migrate Down
DROP INDEX unique_users__email;

