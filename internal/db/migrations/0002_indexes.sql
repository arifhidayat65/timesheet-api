CREATE INDEX IF NOT EXISTS idx_timesheet_period ON timesheets (year, month);
CREATE INDEX IF NOT EXISTS idx_entries_date ON timesheet_entries (work_date);
