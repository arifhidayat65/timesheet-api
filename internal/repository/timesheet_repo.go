package repository

import "timesheet-api/internal/domain"

type Filter struct {
	EmployeeName string
	Month        *int
	Year         *int
}

type TimesheetRepository interface {
	Create(ts *domain.Timesheet) (int64, error)
	FindByID(id int64) (*domain.Timesheet, error)
	List(f Filter) ([]domain.Timesheet, error)
	Update(ts *domain.Timesheet) error
	Delete(id int64) error

	AddEntry(e *domain.TimesheetEntry) (int64, error)
	UpdateEntry(e *domain.TimesheetEntry) error
	DeleteEntry(id int64) error

	// Tambahan untuk summary
	Stats(timesheetID int64) (days int64, totalHours float64, overtimeHours float64, err error)
}
