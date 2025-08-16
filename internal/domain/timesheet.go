package domain

import (
	"errors"
	"time"
)

type Timesheet struct {
	ID               int64
	EmployeeName     string
	Department       string
	Month            int
	Year             int
	TotalWorkingDays *int
	CreatedAt        time.Time
	Entries          []TimesheetEntry
}

type TimesheetEntry struct {
	ID            int64
	TimesheetID   int64
	WorkDate      time.Time
	StartTime     *time.Time
	EndTime       *time.Time
	TotalHours    *float64
	OvertimeHours *float64
	Remarks       string
	CreatedAt     time.Time
}

var (
	ErrNotFound     = errors.New("not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrDuplicate    = errors.New("duplicate")
)
