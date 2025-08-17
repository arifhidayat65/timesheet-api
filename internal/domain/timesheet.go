package domain

import (
	"errors"
	"time"
)

type Timesheet struct {
	ID               int64            `json:"id"`
	EmployeeName     string           `json:"employee_name"`
	Department       string           `json:"department"`
	Month            int              `json:"month"`
	Year             int              `json:"year"`
	TotalWorkingDays *int             `json:"total_working_days,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
	Entries          []TimesheetEntry `json:"entries,omitempty"`
}

type TimesheetEntry struct {
	ID            int64      `json:"id"`
	TimesheetID   int64      `json:"timesheet_id"`
	WorkDate      time.Time  `json:"date"`
	StartTime     *time.Time `json:"start_time,omitempty"`
	EndTime       *time.Time `json:"end_time,omitempty"`
	TotalHours    *float64   `json:"total_hours,omitempty"`
	OvertimeHours *float64   `json:"overtime_hours,omitempty"`
	Remarks       string     `json:"remarks,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

var (
	ErrNotFound     = errors.New("not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrDuplicate    = errors.New("duplicate")
)
