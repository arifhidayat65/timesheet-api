CREATE TABLE IF NOT EXISTS timesheets (
  id BIGSERIAL PRIMARY KEY,
  employee_name VARCHAR(100) NOT NULL,
  department    VARCHAR(100),
  month         SMALLINT NOT NULL CHECK (month BETWEEN 1 AND 12),
  year          SMALLINT NOT NULL CHECK (year BETWEEN 1900 AND 2100),
  total_working_days INT,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (employee_name, month, year)
);

CREATE TABLE IF NOT EXISTS timesheet_entries (
  id BIGSERIAL PRIMARY KEY,
  timesheet_id BIGINT NOT NULL REFERENCES timesheets(id) ON DELETE CASCADE,
  work_date    DATE NOT NULL,
  start_time   TIME,
  end_time     TIME,
  total_hours      NUMERIC(5,2),
  overtime_hours   NUMERIC(5,2),
  remarks      TEXT,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
