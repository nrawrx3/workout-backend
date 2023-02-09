-- +migrate Up
CREATE TABLE workouts (
  id integer PRIMARY KEY,
  created_at datetime,
  updated_at datetime,
  deleted_at datetime,
  kind text,
  reps integer,
  rounds integer,
  duration_seconds integer,
  relative_order integer,
  user_id integer,
  FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- +migrate Down
DROP TABLE workouts;

