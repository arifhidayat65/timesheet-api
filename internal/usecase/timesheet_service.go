package usecase

import (
	"math"
	"time"

	"timesheet-api/internal/domain"
	"timesheet-api/internal/repository"
)

type TimesheetService struct {
	repo repository.TimesheetRepository
}

func NewTimesheetService(r repository.TimesheetRepository) *TimesheetService {
	return &TimesheetService{repo: r}
}

// ---------- Timesheet header ----------

func (s *TimesheetService) CreateTimesheet(ts *domain.Timesheet) (int64, error) {
	if ts.EmployeeName == "" || ts.Month < 1 || ts.Month > 12 || ts.Year < 1900 || ts.Year > 2100 {
		return 0, domain.ErrInvalidInput
	}
	return s.repo.Create(ts)
}

func (s *TimesheetService) GetTimesheet(id int64) (*domain.Timesheet, error) {
	return s.repo.FindByID(id)
}

func (s *TimesheetService) ListTimesheets(f repository.Filter) ([]domain.Timesheet, error) {
	return s.repo.List(f)
}

func (s *TimesheetService) UpdateTimesheet(ts *domain.Timesheet) error {
	if ts.ID == 0 {
		return domain.ErrInvalidInput
	}
	return s.repo.Update(ts)
}

func (s *TimesheetService) DeleteTimesheet(id int64) error {
	return s.repo.Delete(id)
}

// ---------- Entries ----------

func (s *TimesheetService) AddEntry(e *domain.TimesheetEntry) (int64, error) {
	if e.TimesheetID == 0 || e.WorkDate.IsZero() {
		return 0, domain.ErrInvalidInput
	}
	// Auto-calc total hours jika kosong & start/end ada
	if e.TotalHours == nil && e.StartTime != nil && e.EndTime != nil {
		dur := e.EndTime.Sub(*e.StartTime).Hours()
		h := math.Round(dur*100) / 100
		e.TotalHours = &h
	}
	return s.repo.AddEntry(e)
}

func (s *TimesheetService) UpdateEntry(e *domain.TimesheetEntry) error {
	if e.ID == 0 {
		return domain.ErrInvalidInput
	}
	if e.TotalHours == nil && e.StartTime != nil && e.EndTime != nil {
		dur := e.EndTime.Sub(*e.StartTime).Hours()
		h := math.Round(dur*100) / 100
		e.TotalHours = &h
	}
	return s.repo.UpdateEntry(e)
}

func (s *TimesheetService) DeleteEntry(id int64) error {
	return s.repo.DeleteEntry(id)
}

// Helpers to build time from string (HH:MM or HH:MM:SS)
func ParseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

func ParseTime(s string) (*time.Time, error) {
	if s == "" { return nil, nil }
	if t, err := time.Parse("15:04:05", s); err == nil { return &t, nil }
	if t, err := time.Parse("15:04", s); err == nil {
		ts, _ := time.Parse("15:04:05", t.Format("15:04")+":00")
		return &ts, nil
	}
	return nil, domain.ErrInvalidInput
}
