package resp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type PaginationMeta struct {
	Page     int   `json:"page,omitempty"`
	PageSize int   `json:"page_size,omitempty"`
	Total    int64 `json:"total,omitempty"`
}

type ResponseMeta struct {
	RequestID  string           `json:"request_id,omitempty"`
	Pagination *PaginationMeta  `json:"pagination,omitempty"`
}

type ErrorDetail struct {
	Type    string `json:"type,omitempty"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message,omitempty"`
}

type APIResponse struct {
	Success bool          `json:"success"`
	Code    int           `json:"code"`
	Status  string        `json:"status"`
	Message string        `json:"message,omitempty"`
	Data    interface{}   `json:"data,omitempty"`
	Error   interface{}   `json:"error,omitempty"`
	Meta    *ResponseMeta `json:"meta,omitempty"`
}

func write(c *gin.Context, code int, status, msg string, data interface{}, errDetail interface{}, meta *ResponseMeta) {
	c.JSON(code, APIResponse{
		Success: code >= 200 && code < 300,
		Code:    code,
		Status:  status,
		Message: msg,
		Data:    data,
		Error:   errDetail,
		Meta:    meta,
	})
}

func OK(c *gin.Context, data interface{}, msg string) {
	write(c, http.StatusOK, http.StatusText(http.StatusOK), msg, data, nil, nil)
}
func Created(c *gin.Context, data interface{}, msg string) {
	write(c, http.StatusCreated, http.StatusText(http.StatusCreated), msg, data, nil, nil)
}
func NoContent(c *gin.Context) {
	write(c, http.StatusNoContent, http.StatusText(http.StatusNoContent), "No Content", nil, nil, nil)
}

func BadRequest(c *gin.Context, err interface{}, msg string) {
	write(c, http.StatusBadRequest, http.StatusText(http.StatusBadRequest), msg, nil, err, nil)
}
func Unauthorized(c *gin.Context, msg string) {
	write(c, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), msg, nil, nil, nil)
}
func Forbidden(c *gin.Context, msg string) {
	write(c, http.StatusForbidden, http.StatusText(http.StatusForbidden), msg, nil, nil, nil)
}
func NotFound(c *gin.Context, msg string) {
	write(c, http.StatusNotFound, http.StatusText(http.StatusNotFound), msg, nil, nil, nil)
}
func Conflict(c *gin.Context, msg string) {
	write(c, http.StatusConflict, http.StatusText(http.StatusConflict), msg, nil, nil, nil)
}
func Unprocessable(c *gin.Context, errs []ErrorDetail, msg string) {
	write(c, http.StatusUnprocessableEntity, http.StatusText(http.StatusUnprocessableEntity), msg, nil, errs, nil)
}
func TooManyRequests(c *gin.Context, msg string) {
	write(c, http.StatusTooManyRequests, http.StatusText(http.StatusTooManyRequests), msg, nil, nil, nil)
}

func Internal(c *gin.Context, msg string) {
	write(c, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), msg, nil, nil, nil)
}
func ServiceUnavailable(c *gin.Context, msg string) {
	write(c, http.StatusServiceUnavailable, http.StatusText(http.StatusServiceUnavailable), msg, nil, nil, nil)
}
