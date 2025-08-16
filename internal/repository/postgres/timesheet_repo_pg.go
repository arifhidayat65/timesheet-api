package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"timesheet-api/internal/domain"
	"timesheet-api/internal/repository"
)

type TimesheetRepoPG struct {
	DB *sql.DB
}

func NewTimesheetRepoPG(db *sql.DB) *TimesheetRepoPG { return &TimesheetRepoPG{DB: db} }

func (r *TimesheetRepoPG) Create(ts *domain.Timesheet) (int64, error) {
	q := `INSERT INTO timesheets (employee_name, department, month, year, total_working_days)
	      VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at`
	var id int64
	var created time.Time
	err := r.DB.QueryRow(q, ts.EmployeeName, ts.Department, ts.Month, ts.Year, ts.TotalWorkingDays).
		Scan(&id, &created)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return 0, domain.ErrDuplicate
		}
		return 0, err
	}
	ts.ID = id
	ts.CreatedAt = created
	return id, nil
}

func (r *TimesheetRepoPG) FindByID(id int64) (*domain.Timesheet, error) {
	var ts domain.Timesheet
	q := `SELECT id, employee_name, department, month, year, total_working_days, created_at
	      FROM timesheets WHERE id=$1`
	err := r.DB.QueryRow(q, id).
		Scan(&ts.ID, &ts.EmployeeName, &ts.Department, &ts.Month, &ts.Year, &ts.TotalWorkingDays, &ts.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	rows, err := r.DB.Query(`SELECT id, work_date, start_time, end_time, total_hours, overtime_hours, remarks, created_at
	                         FROM timesheet_entries WHERE timesheet_id=$1 ORDER BY work_date ASC`, id)
	if err != nil { return nil, err }
	defer rows.Close()

	var entries []domain.TimesheetEntry
	for rows.Next() {
		var e domain.TimesheetEntry
		var st, et sql.NullTime
		err := rows.Scan(&e.ID, &e.WorkDate, &st, &et, &e.TotalHours, &e.OvertimeHours, &e.Remarks, &e.CreatedAt)
		if err != nil { return nil, err }
		e.TimesheetID = id
		if st.Valid { t := st.Time; e.StartTime = &t }
		if et.Valid { t := et.Time; e.EndTime = &t }
		entries = append(entries, e)
	}
	ts.Entries = entries
	return &ts, nil
}

func (r *TimesheetRepoPG) List(f repository.Filter) ([]domain.Timesheet, error) {
	q := `SELECT id, employee_name, department, month, year, total_working_days, created_at
	      FROM timesheets WHERE 1=1`
	var args []interface{}
	i := 1
	if f.EmployeeName != "" { q += fmt.Sprintf(" AND employee_name = $%d", i); args = append(args, f.EmployeeName); i++ }
	if f.Month != nil { q += fmt.Sprintf(" AND month = $%d", i); args = append(args, *f.Month); i++ }
	if f.Year  != nil { q += fmt.Sprintf(" AND year = $%d", i);  args = append(args, *f.Year);  i++ }
	q += " ORDER BY year DESC, month DESC, id DESC"

	rows, err := r.DB.Query(q, args...)
	if err != nil { return nil, err }
	defer rows.Close()

	var out []domain.Timesheet
	for rows.Next() {
		var t domain.Timesheet
		if err := rows.Scan(&t.ID, &t.EmployeeName, &t.Department, &t.Month, &t.Year, &t.TotalWorkingDays, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, nil
}

func (r *TimesheetRepoPG) Update(ts *domain.Timesheet) error {
	res, err := r.DB.Exec(`UPDATE timesheets SET employee_name=$1, department=$2, month=$3, year=$4, total_working_days=$5 WHERE id=$6`,
		ts.EmployeeName, ts.Department, ts.Month, ts.Year, ts.TotalWorkingDays, ts.ID)
	if err != nil { return err }
	aff, _ := res.RowsAffected()
	if aff == 0 { return domain.ErrNotFound }
	return nil
}

func (r *TimesheetRepoPG) Delete(id int64) error {
	res, err := r.DB.Exec(`DELETE FROM timesheets WHERE id=$1`, id)
	if err != nil { return err }
	aff, _ := res.RowsAffected()
	if aff == 0 { return domain.ErrNotFound }
	return nil
}

func (r *TimesheetRepoPG) AddEntry(e *domain.TimesheetEntry) (int64, error) {
	q := `INSERT INTO timesheet_entries (timesheet_id, work_date, start_time, end_time, total_hours, overtime_hours, remarks)
	      VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id, created_at`
	var id int64
	var created time.Time
	err := r.DB.QueryRow(q, e.TimesheetID, e.WorkDate, e.StartTime, e.EndTime, e.TotalHours, e.OvertimeHours, e.Remarks).
		Scan(&id, &created)
	if err != nil { return 0, err }
	e.ID = id
	e.CreatedAt = created
	return id, nil
}

func (r *TimesheetRepoPG) UpdateEntry(e *domain.TimesheetEntry) error {
	q := `UPDATE timesheet_entries
	      SET work_date=$1, start_time=$2, end_time=$3, total_hours=$4, overtime_hours=$5, remarks=$6
	      WHERE id=$7`
	res, err := r.DB.Exec(q, e.WorkDate, e.StartTime, e.EndTime, e.TotalHours, e.OvertimeHours, e.Remarks, e.ID)
	if err != nil { return err }
	aff, _ := res.RowsAffected()
	if aff == 0 { return domain.ErrNotFound }
	return nil
}

func (r *TimesheetRepoPG) DeleteEntry(id int64) error {
	res, err := r.DB.Exec(`DELETE FROM timesheet_entries WHERE id=$1`, id)
	if err != nil { return err }
	aff, _ := res.RowsAffected()
	if aff == 0 { return domain.ErrNotFound }
	return nil
}

// Optional: transaction template for future batch ops
func (r *TimesheetRepoPG) withTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil { return err }
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
