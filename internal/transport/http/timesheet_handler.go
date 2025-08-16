package http

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"timesheet-api/internal/domain"
	"timesheet-api/internal/repository"
	"timesheet-api/internal/resp"
	"timesheet-api/internal/usecase"
)

type TimesheetHandler struct {
	svc *usecase.TimesheetService
}

func NewTimesheetHandler(svc *usecase.TimesheetService) *TimesheetHandler { return &TimesheetHandler{svc: svc} }

func (h *TimesheetHandler) Register(r *gin.Engine) {
	r.POST("/timesheets", h.createTimesheet)
	r.GET("/timesheets", h.listTimesheets)
	r.GET("/timesheets/:id", h.getTimesheet)
	r.PUT("/timesheets/:id", h.updateTimesheet)
	r.DELETE("/timesheets/:id", h.deleteTimesheet)

	r.POST("/timesheets/:id/entries", h.addEntry)
	r.PUT("/timesheets/:tsid/entries/:id", h.updateEntry)
	r.DELETE("/timesheets/:tsid/entries/:id", h.deleteEntry)
}

// ---------- DTOs ----------

type createTimesheetReq struct {
	EmployeeName     string `json:"employee_name" binding:"required"`
	Department       string `json:"department"`
	Month            int    `json:"month" binding:"required"`
	Year             int    `json:"year" binding:"required"`
	TotalWorkingDays *int   `json:"total_working_days"`
}

type updateTimesheetReq createTimesheetReq

type entryReq struct {
	Date          string   `json:"date" binding:"required"`       // YYYY-MM-DD
	StartTime     string   `json:"start_time"`                     // HH:MM or HH:MM:SS (accept "08.00" too)
	EndTime       string   `json:"end_time"`
	TotalHours    *float64 `json:"total_hours"`
	OvertimeHours *float64 `json:"overtime_hours"`
	Remarks       string   `json:"remarks"`
}

// ---------- Handlers ----------

func (h *TimesheetHandler) createTimesheet(c *gin.Context) {
	var req createTimesheetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Unprocessable(c, []resp.ErrorDetail{{Type:"validation_error", Field:"body", Message: err.Error()}}, "Invalid payload")
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
	if err != nil {
		h.mapError(c, err)
		return
	}
	resp.Created(c, gin.H{"id": id}, "Timesheet created")
}

func (h *TimesheetHandler) listTimesheets(c *gin.Context) {
	name := c.Query("employee_name")
	var mptr, yptr *int
	if v := c.Query("month"); v != "" { if n, err := strconv.Atoi(v); err==nil { mptr=&n } }
	if v := c.Query("year"); v != ""  { if n, err := strconv.Atoi(v); err==nil { yptr=&n } }
	items, err := h.svc.ListTimesheets(repository.Filter{
		EmployeeName: name,
		Month: mptr,
		Year:  yptr,
	})
	if err != nil { h.mapError(c, err); return }
	resp.OK(c, items, "Success")
}

func (h *TimesheetHandler) getTimesheet(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ts, err := h.svc.GetTimesheet(id)
	if err != nil { h.mapError(c, err); return }
	resp.OK(c, ts, "Success")
}

func (h *TimesheetHandler) updateTimesheet(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req updateTimesheetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Unprocessable(c, []resp.ErrorDetail{{Type:"validation_error", Field:"body", Message: err.Error()}}, "Invalid payload")
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
	tsID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req entryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Unprocessable(c, []resp.ErrorDetail{{Type:"validation_error", Field:"body", Message: err.Error()}}, "Invalid payload")
		return
	}
	// Normalize time allowing "08.00"
	st := strings.ReplaceAll(req.StartTime, ".", ":")
	et := strings.ReplaceAll(req.EndTime, ".", ":")

	d, err := usecase.ParseDate(req.Date)
	if err != nil {
		resp.BadRequest(c, []resp.ErrorDetail{{Type:"validation_error", Field:"date", Message:"format YYYY-MM-DD"}}, "Invalid date")
		return
	}
	stp, err := usecase.ParseTime(st)
	if err != nil && st != "" {
		resp.BadRequest(c, []resp.ErrorDetail{{Type:"validation_error", Field:"start_time", Message:"format HH:MM atau HH:MM:SS"}}, "Invalid start_time")
		return
	}
	etp, err := usecase.ParseTime(et)
	if err != nil && et != "" {
		resp.BadRequest(c, []resp.ErrorDetail{{Type:"validation_error", Field:"end_time", Message:"format HH:MM atau HH:MM:SS"}}, "Invalid end_time")
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
		resp.Unprocessable(c, []resp.ErrorDetail{{Type:"validation_error", Field:"body", Message: err.Error()}}, "Invalid payload")
		return
	}
	var d time.Time
	var err error
	if req.Date != "" {
		d, err = usecase.ParseDate(req.Date)
		if err != nil {
			resp.BadRequest(c, []resp.ErrorDetail{{Type:"validation_error", Field:"date", Message:"format YYYY-MM-DD"}}, "Invalid date")
			return
		}
	}
	st := strings.ReplaceAll(req.StartTime, ".", ":")
	et := strings.ReplaceAll(req.EndTime, ".", ":")
	stp, err := usecase.ParseTime(st); if err != nil && st != "" {
		resp.BadRequest(c, []resp.ErrorDetail{{Type:"validation_error", Field:"start_time", Message:"format HH:MM atau HH:MM:SS"}}, "Invalid start_time")
		return
	}
	etp, err := usecase.ParseTime(et); if err != nil && et != "" {
		resp.BadRequest(c, []resp.ErrorDetail{{Type:"validation_error", Field:"end_time", Message:"format HH:MM atau HH:MM:SS"}}, "Invalid end_time")
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
