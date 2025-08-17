package http

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"

	"timesheet-api/internal/domain"
	"timesheet-api/internal/repository"
	"timesheet-api/internal/resp"
	"timesheet-api/internal/usecase"
)

type TimesheetHandler struct{ svc *usecase.TimesheetService }
func NewTimesheetHandler(s *usecase.TimesheetService) *TimesheetHandler { return &TimesheetHandler{svc: s} }

func (h *TimesheetHandler) Register(r *gin.Engine) {
	ts := r.Group("/timesheets")
	{
		ts.POST("", h.createTimesheet)
		ts.GET("", h.listTimesheets)
		ts.GET("/:id", h.getTimesheet)
		ts.PUT("/:id", h.updateTimesheet)
		ts.DELETE("/:id", h.deleteTimesheet)
		// JANGAN buat apa pun di bawah ts.POST("/:id/...") — itu yang bikin panic
	}

	entries := r.Group("/entries")
	{
		entries.POST("", h.addEntry)       // ?timesheet_id=...
		entries.PUT("/:id", h.updateEntry)
		entries.DELETE("/:id", h.deleteEntry)
	}
}

// ====== Request/Response DTO ======

type createTimesheetReq struct {
	EmployeeName     string `json:"employee_name" binding:"required"`
	Department       string `json:"department"`
	Month            int    `json:"month" binding:"required"`
	Year             int    `json:"year" binding:"required"`
	TotalWorkingDays *int   `json:"total_working_days"`
}
type updateTimesheetReq createTimesheetReq

type entryReq struct {
	Date          string   `json:"date" binding:"required"`
	StartTime     string   `json:"start_time"`
	EndTime       string   `json:"end_time"`
	TotalHours    *float64 `json:"total_hours"`
	OvertimeHours *float64 `json:"overtime_hours"`
	Remarks       string   `json:"remarks"`
}

type entryResponse struct {
	ID            int64    `json:"id"`
	Date          string   `json:"date"`
	DayName       string   `json:"day_name"`
	StartTime     *string  `json:"start_time,omitempty"`
	EndTime       *string  `json:"end_time,omitempty"`
	TotalHours    *float64 `json:"total_hours,omitempty"`
	OvertimeHours *float64 `json:"overtime_hours,omitempty"`
	Remarks       string   `json:"remarks,omitempty"`
}
type timesheetResponse struct {
	ID               int64           `json:"id"`
	EmployeeName     string          `json:"employee_name"`
	Department       string          `json:"department"`
	Month            int             `json:"month"`
	Year             int             `json:"year"`
	TotalWorkingDays *int            `json:"total_working_days,omitempty"`
	Summary          struct {
		DaysFilled    int64   `json:"days_filled"`
		TotalHours    float64 `json:"total_hours"`
		OvertimeHours float64 `json:"overtime_hours"`
	} `json:"summary"`
	Entries []entryResponse `json:"entries"`
}

// ====== Handlers ======

func (h *TimesheetHandler) createTimesheet(c *gin.Context) {
	var req createTimesheetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Unprocessable(c, []resp.ErrorDetail{{Type: "validation_error", Field: "body", Message: err.Error()}}, "Invalid payload")
		return
	}
	ts := domain.Timesheet{
		EmployeeName:     req.EmployeeName,
		Department:       req.Department,
		Month:            req.Month,
		Year:             req.Year,
		TotalWorkingDays: req.TotalWorkingDays,
	}
	id, err := h.svc.CreateTimesheet(&ts)
	if err != nil { h.mapError(c, err); return }
	resp.Created(c, gin.H{"id": id}, "Timesheet created")
}

func (h *TimesheetHandler) listTimesheets(c *gin.Context) {
	name := c.Query("employee_name")
	var mptr, yptr *int
	if v := c.Query("month"); v != "" { if n, err := strconv.Atoi(v); err == nil { mptr = &n } }
	if v := c.Query("year");  v != "" { if n, err := strconv.Atoi(v); err == nil { yptr = &n } }
	items, err := h.svc.ListTimesheets(repository.Filter{EmployeeName: name, Month: mptr, Year: yptr})
	if err != nil { h.mapError(c, err); return }
	resp.OK(c, items, "Success")
}

func (h *TimesheetHandler) getTimesheet(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ts, err := h.svc.GetTimesheet(id)
	if err != nil { h.mapError(c, err); return }

	// Summary
	days, th, oh, err := h.svc.Stats(id)
	if err != nil { h.mapError(c, err); return }

	// Map entries + day name
	ers := make([]entryResponse, 0, len(ts.Entries))
	for _, e := range ts.Entries {
		var st, et *string
		if e.StartTime != nil { s := e.StartTime.Format("15:04:05"); st = &s }
		if e.EndTime   != nil { s := e.EndTime.Format("15:04:05");   et = &s }
		dayName := indoDayName(e.WorkDate.Weekday())
		ers = append(ers, entryResponse{
			ID: e.ID, Date: e.WorkDate.Format("2006-01-02"), DayName: dayName,
			StartTime: st, EndTime: et, TotalHours: e.TotalHours, OvertimeHours: e.OvertimeHours, Remarks: e.Remarks,
		})
	}
	out := timesheetResponse{
		ID: ts.ID, EmployeeName: ts.EmployeeName, Department: ts.Department,
		Month: ts.Month, Year: ts.Year, TotalWorkingDays: ts.TotalWorkingDays, Entries: ers,
	}
	out.Summary.DaysFilled = days
	out.Summary.TotalHours = th
	out.Summary.OvertimeHours = oh
	resp.OK(c, out, "Success")
}

func (h *TimesheetHandler) updateTimesheet(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req updateTimesheetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Unprocessable(c, []resp.ErrorDetail{{Type: "validation_error", Field: "body", Message: err.Error()}}, "Invalid payload")
		return
	}
	ts := domain.Timesheet{
		ID:               id,
		EmployeeName:     req.EmployeeName,
		Department:       req.Department,
		Month:            req.Month,
		Year:             req.Year,
		TotalWorkingDays: req.TotalWorkingDays,
	}
	if err := h.svc.UpdateTimesheet(&ts); err != nil { h.mapError(c, err); return }
	resp.OK(c, gin.H{"id": id}, "Timesheet updated")
}

func (h *TimesheetHandler) deleteTimesheet(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.svc.DeleteTimesheet(id); err != nil { h.mapError(c, err); return }
	resp.NoContent(c)
}

func (h *TimesheetHandler) addEntry(c *gin.Context) {
	// Ambil timesheet_id dari QUERY (bukan nested route)
	tsIDStr := c.Query("timesheet_id")
	if tsIDStr == "" {
		resp.BadRequest(c, []resp.ErrorDetail{{Type: "validation_error", Field: "timesheet_id", Message: "wajib diisi (query param)"}}, "Missing timesheet_id")
		return
	}
	tsID, err := strconv.ParseInt(tsIDStr, 10, 64)
	if err != nil || tsID <= 0 {
		resp.BadRequest(c, []resp.ErrorDetail{{Type: "validation_error", Field: "timesheet_id", Message: "harus angka > 0"}}, "Invalid timesheet_id")
		return
	}

	var req entryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Unprocessable(c, []resp.ErrorDetail{{Type: "validation_error", Field: "body", Message: err.Error()}}, "Invalid payload")
		return
	}
	st := strings.ReplaceAll(req.StartTime, ".", ":")
	et := strings.ReplaceAll(req.EndTime, ".", ":")

	d, err := usecase.ParseDate(req.Date)
	if err != nil {
		resp.BadRequest(c, []resp.ErrorDetail{{Type: "validation_error", Field: "date", Message: "format YYYY-MM-DD"}}, "Invalid date")
		return
	}
	stp, err := usecase.ParseTime(st); if err != nil && st != "" {
		resp.BadRequest(c, []resp.ErrorDetail{{Type: "validation_error", Field: "start_time", Message: "format HH:MM atau HH:MM:SS"}}, "Invalid start_time")
		return
	}
	etp, err := usecase.ParseTime(et); if err != nil && et != "" {
		resp.BadRequest(c, []resp.ErrorDetail{{Type: "validation_error", Field: "end_time", Message: "format HH:MM atau HH:MM:SS"}}, "Invalid end_time")
		return
	}

	e := domain.TimesheetEntry{
		TimesheetID:   tsID,
		WorkDate:      d,
		StartTime:     stp,
		EndTime:       etp,
		TotalHours:    req.TotalHours,
		OvertimeHours: req.OvertimeHours,
		Remarks:       req.Remarks,
	}
	id, err := h.svc.AddEntry(&e)
	if err != nil { h.mapError(c, err); return }
	resp.Created(c, gin.H{"id": id}, "Entry created")
}

func (h *TimesheetHandler) updateEntry(c *gin.Context) {
	entryID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req entryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Unprocessable(c, []resp.ErrorDetail{{Type: "validation_error", Field: "body", Message: err.Error()}}, "Invalid payload")
		return
	}
	var d time.Time
	var err error
	if req.Date != "" {
		d, err = usecase.ParseDate(req.Date)
		if err != nil {
			resp.BadRequest(c, []resp.ErrorDetail{{Type: "validation_error", Field: "date", Message: "format YYYY-MM-DD"}}, "Invalid date")
			return
		}
	}
	st := strings.ReplaceAll(req.StartTime, ".", ":")
	et := strings.ReplaceAll(req.EndTime, ".", ":")
	stp, err := usecase.ParseTime(st); if err != nil && st != "" {
		resp.BadRequest(c, []resp.ErrorDetail{{Type: "validation_error", Field: "start_time", Message: "format HH:MM atau HH:MM:SS"}}, "Invalid start_time")
		return
	}
	etp, err := usecase.ParseTime(et); if err != nil && et != "" {
		resp.BadRequest(c, []resp.ErrorDetail{{Type: "validation_error", Field: "end_time", Message: "format HH:MM atau HH:MM:SS"}}, "Invalid end_time")
		return
	}

	e := domain.TimesheetEntry{
		ID:            entryID,
		WorkDate:      d,
		StartTime:     stp,
		EndTime:       etp,
		TotalHours:    req.TotalHours,
		OvertimeHours: req.OvertimeHours,
		Remarks:       req.Remarks,
	}
	if err := h.svc.UpdateEntry(&e); err != nil { h.mapError(c, err); return }
	resp.OK(c, gin.H{"id": entryID}, "Entry updated")
}

func (h *TimesheetHandler) deleteEntry(c *gin.Context) {
	entryID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.svc.DeleteEntry(entryID); err != nil { h.mapError(c, err); return }
	resp.NoContent(c)
}

// ====== Export PDF ======

func (h *TimesheetHandler) exportTimesheetPDF(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ts, err := h.svc.GetTimesheet(id)
	if err != nil { h.mapError(c, err); return }

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 12, 10)
	pdf.AddPage()
	pdf.SetAutoPageBreak(true, 10)

	// Title
	pdf.SetFont("Helvetica", "B", 16)
	pdf.Cell(0, 8, "ABSENSI KEHADIRAN / TIME SHEET")
	pdf.Ln(10)

	// Header info
	pdf.SetFont("Helvetica", "", 11)
	headerRow := func(label, val string) {
		pdf.CellFormat(55, 6, label, "", 0, "", false, 0, "")
		pdf.CellFormat(5, 6, ":", "", 0, "", false, 0, "")
		pdf.CellFormat(0, 6, val, "", 1, "", false, 0, "")
	}
	headerRow("Nama Karyawan / Employee Name", ts.EmployeeName)
	headerRow("Divisi / Department", ts.Department)
	headerRow("Periode", fmt.Sprintf("%s %d", indoMonth(ts.Month), ts.Year))
	if ts.TotalWorkingDays != nil {
		headerRow("Total Hari Kerja / Total Working Day", fmt.Sprintf("%d Hari", *ts.TotalWorkingDays))
	}
	pdf.Ln(2)

	cols := []struct {
		Title string
		Width float64
	}{
		{"Tanggal / Date", 25},
		{"Hari / Day", 22},
		{"Mulai / Start", 22},
		{"Selesai / End", 22},
		{"Jam / Hrs", 20},
		{"Lembur / OT", 22},
		{"Keterangan / Remarks", 57},
	}

	// Header tabel
	pdf.SetFont("Helvetica", "B", 10)
	for _, col := range cols {
		pdf.CellFormat(col.Width, 8, col.Title, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)

	// Body
	pdf.SetFont("Helvetica", "", 10)
	var totalHrs, totalOT float64
	for _, e := range ts.Entries {
		date := e.WorkDate.Format("2006-01-02")
		day := indoDayName(e.WorkDate.Weekday())
		var st, et string
		if e.StartTime != nil { st = e.StartTime.Format("15:04") } else { st = "-" }
		if e.EndTime != nil   { et = e.EndTime.Format("15:04")   } else { et = "-" }
		var th, ot string
		if e.TotalHours != nil { th = fmt.Sprintf("%.2f", *e.TotalHours); totalHrs += *e.TotalHours } else { th = "-" }
		if e.OvertimeHours != nil { ot = fmt.Sprintf("%.2f", *e.OvertimeHours); totalOT += *e.OvertimeHours } else { ot = "-" }
		remarks := e.Remarks

		pdf.CellFormat(cols[0].Width, 8, date, "1", 0, "C", false, 0, "")
		pdf.CellFormat(cols[1].Width, 8, day,  "1", 0, "C", false, 0, "")
		pdf.CellFormat(cols[2].Width, 8, st,   "1", 0, "C", false, 0, "")
		pdf.CellFormat(cols[3].Width, 8, et,   "1", 0, "C", false, 0, "")
		pdf.CellFormat(cols[4].Width, 8, th,   "1", 0, "C", false, 0, "")
		pdf.CellFormat(cols[5].Width, 8, ot,   "1", 0, "C", false, 0, "")
		pdf.MultiCell(cols[6].Width, 8, remarks, "1", "L", false)
	}

	// Total
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(cols[0].Width+cols[1].Width+cols[2].Width+cols[3].Width, 8, "TOTAL", "1", 0, "R", false, 0, "")
	pdf.CellFormat(cols[4].Width, 8, fmt.Sprintf("%.2f", totalHrs), "1", 0, "C", false, 0, "")
	pdf.CellFormat(cols[5].Width, 8, fmt.Sprintf("%.2f", totalOT),  "1", 0, "C", false, 0, "")
	pdf.CellFormat(cols[6].Width, 8, "", "1", 1, "L", false, 0, "")

	var b bytes.Buffer
	if err := pdf.Output(&b); err != nil {
		resp.Internal(c, "gagal membuat PDF")
		return
	}
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=timesheet_%d_%02d_%d.pdf", ts.Year, ts.Month, ts.ID))
	c.Data(200, "application/pdf", b.Bytes())
}

// ====== Helpers ======

func (h *TimesheetHandler) mapError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		resp.BadRequest(c, fmt.Sprintf("%v", err), "Invalid input")
	case errors.Is(err, domain.ErrNotFound):
		resp.NotFound(c, "Not found")
	case errors.Is(err, domain.ErrDuplicate):
		resp.Conflict(c, "Duplicate")
	default:
		resp.Internal(c, "Internal error")
	}
}

func indoMonth(m int) string {
	names := []string{"", "Januari","Februari","Maret","April","Mei","Juni","Juli","Agustus","September","Oktober","November","Desember"}
	if m >= 1 && m <= 12 { return names[m] }
	return fmt.Sprintf("Bulan-%d", m)
}
func indoDayName(w time.Weekday) string {
	switch w {
	case time.Monday: return "Senin"
	case time.Tuesday: return "Selasa"
	case time.Wednesday: return "Rabu"
	case time.Thursday: return "Kamis"
	case time.Friday: return "Jumat"
	case time.Saturday: return "Sabtu"
	default: return "Minggu"
	}
}
